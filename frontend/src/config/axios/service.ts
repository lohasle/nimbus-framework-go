import axios, { AxiosError, AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'

import { ElMessage, ElNotification } from 'element-plus'
import qs from 'qs'
import { config } from '@/config/axios/config'
import {
  getAccessToken,
  getRefreshToken,
  getTenantId,
  getVisitTenantId,
  removeToken,
  setToken
} from '@/utils/auth'
import errorCode from './errorCode'

import router, { resetRouter } from '@/router'
import { deleteUserCache } from '@/hooks/web/useCache'
import { ApiEncrypt } from '@/utils/encrypt'

const tenantEnable = import.meta.env.VITE_APP_TENANT_ENABLE
const { result_code, base_url, request_timeout } = config

interface AuthRequestConfig extends InternalAxiosRequestConfig {
  __authRetry?: boolean
  __skipAuthRefresh?: boolean
}

// 同一页面内所有 401 共享一次刷新；失败退登也只能执行一次。
let refreshPromise: Promise<string> | null = null
let sessionExpiryStarted = false
// 请求白名单：这些公开接口的 401 属于业务校验失败，不能触发会话过期流程。
const whiteList: string[] = [
  '/system/auth/login',
  '/system/auth/register',
  '/system/auth/refresh-token',
  '/system/auth/send-sms-code',
  '/system/auth/sms-login',
  '/system/auth/social-login',
  '/system/auth/social-auth-redirect',
  '/system/auth/reset-password',
  '/system/tenant/get-id-by-name',
  '/system/tenant/get-by-website',
  'system/captcha/'
]

// 创建axios实例
const service: AxiosInstance = axios.create({
  baseURL: base_url, // api 的 base_url
  timeout: request_timeout, // 请求超时时间
  withCredentials: false, // 禁用 Cookie 等信息
  // 自定义参数序列化函数
  paramsSerializer: (params) => {
    return qs.stringify(params, { allowDots: true })
  }
})

// request拦截器
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // 是否需要设置 token；命中白名单的接口（如 /login）不带 token
    let isToken = (config!.headers || {}).isToken !== false
    if (isToken && whiteList.some((v) => config.url?.includes(v))) {
      isToken = false
    }
    ;(config as AuthRequestConfig).__skipAuthRefresh = !isToken
    if (getAccessToken() && isToken) {
      config.headers.Authorization = 'Bearer ' + getAccessToken() // 让每个请求携带自定义 token
    }
    // 设置租户
    if (tenantEnable && tenantEnable === 'true') {
      const tenantId = getTenantId()
      if (tenantId) config.headers['tenant-id'] = tenantId
      // 只有登录时，才设置 visit-tenant-id 访问租户
      const visitTenantId = getVisitTenantId()
      if (config.headers.Authorization && visitTenantId) {
        config.headers['visit-tenant-id'] = visitTenantId
      }
    }
    const method = config.method?.toUpperCase()
    // 防止 GET 请求缓存
    if (method === 'GET') {
      config.headers['Cache-Control'] = 'no-cache'
      config.headers['Pragma'] = 'no-cache'
    }
    // 自定义参数序列化函数
    else if (method === 'POST') {
      const contentType = config.headers['Content-Type'] || config.headers['content-type']
      if (contentType === 'application/x-www-form-urlencoded') {
        if (config.data && typeof config.data !== 'string') {
          config.data = qs.stringify(config.data)
        }
      }
    }
    // 是否 API 加密
    if ((config!.headers || {}).isEncrypt && !(config!.headers || {}).isEncrypted) {
      try {
        // 加密请求数据
        if (config.data) {
          config.data = ApiEncrypt.encryptRequest(config.data)
          // 设置加密标识头
          config.headers[ApiEncrypt.getEncryptHeader()] = 'true'
        }
      } catch (error) {
        console.error('请求数据加密失败:', error)
        throw error
      }
    }
    return config
  },
  (error: AxiosError) => {
    // Do something with request error
    console.log(error) // for debug
    return Promise.reject(error)
  }
)

// response 拦截器
service.interceptors.response.use(
  async (response: AxiosResponse<any>) => {
    let { data } = response
    const config = response.config as AuthRequestConfig
    if (!data) {
      // 返回“[HTTP]请求没有返回值”;
      throw new Error()
    }

    // 检查是否需要解密响应数据
    const encryptHeader = ApiEncrypt.getEncryptHeader()
    const isEncryptResponse =
      response.headers[encryptHeader] === 'true' ||
      response.headers[encryptHeader.toLowerCase()] === 'true'
    if (isEncryptResponse && typeof data === 'string') {
      try {
        // 解密响应数据
        data = ApiEncrypt.decryptResponse(data)
      } catch (error) {
        console.error('响应数据解密失败:', error)
        throw new Error('响应数据解密失败: ' + (error as Error).message)
      }
    }

    const { t } = useI18n()
    // 未设置状态码则默认成功状态
    // 二进制数据则直接返回，例如说 Excel 导出
    if (
      response.request.responseType === 'blob' ||
      response.request.responseType === 'arraybuffer'
    ) {
      // 注意：如果导出的响应为 json，说明可能失败了，不直接返回进行下载
      if (response.data.type !== 'application/json') {
        return response.data
      }
      data = await new Response(response.data).json()
    }
    const code = data.code ?? result_code
    // 获取错误信息
    const msg = data.msg || errorCode[code] || errorCode['default']
    if (code === 401) {
      if (config.__skipAuthRefresh) {
        ElNotification.error({ title: msg })
        return Promise.reject(new Error(msg))
      }
      return handleUnauthorized(config)
    } else if (code === 500) {
      ElMessage.error(t('sys.api.errMsg500'))
      return Promise.reject(new Error(msg))
    } else if (code === 901) {
      ElMessage.error({
        offset: 300,
        dangerouslyUseHTMLString: true,
        message:
          '<div>' +
          t('sys.api.errMsg901') +
          '</div>' +
          '<div> &nbsp; </div>' +
          '<div>请查看Nimbus Framework本地开发文档</div>' +
          '<div> &nbsp; </div>' +
          '<div>5 分钟搭建本地环境</div>'
      })
      return Promise.reject(new Error(msg))
    } else if (code !== 0 && code !== 200) {
      ElNotification.error({ title: msg })
      return Promise.reject('error')
    } else {
      return data
    }
  },
  async (error: AxiosError) => {
    const requestConfig = error.config as AuthRequestConfig | undefined
    if (error.response?.status === 401 && requestConfig && !requestConfig.__skipAuthRefresh) {
      return handleUnauthorized(requestConfig)
    }
    console.log('err' + error) // for debug
    let message = (error.response?.data as any)?.msg || error.message
    const { t } = useI18n()
    if (message === 'Network Error') {
      message = t('sys.api.errorMessage')
    } else if (message.includes('timeout')) {
      message = t('sys.api.apiTimeoutMessage')
    } else if (message.includes('Request failed with status code')) {
      message = t('sys.api.apiRequestFailed') + message.substr(message.length - 3)
    }
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

const refreshToken = async () => {
  return await axios.post(
    base_url + '/system/auth/refresh-token?refreshToken=' + getRefreshToken(),
    undefined,
    { headers: { 'tenant-id': getTenantId() } }
  )
}

const refreshAccessTokenWithLock = async (expiredAccessToken?: string): Promise<string> => {
  const executeRefresh = async (): Promise<string> => {
    const latestAccessToken = getAccessToken()
    if (expiredAccessToken && latestAccessToken && latestAccessToken !== expiredAccessToken) {
      return latestAccessToken
    }

    if (!getRefreshToken()) {
      throw new Error('刷新令牌不存在')
    }
    const refreshTokenRes = await refreshToken()
    setToken(refreshTokenRes.data.data)
    const refreshedAccessToken = getAccessToken()
    if (!refreshedAccessToken) {
      throw new Error('刷新访问令牌失败')
    }
    return refreshedAccessToken
  }

  if (typeof navigator !== 'undefined' && navigator.locks) {
    const tenantId = getTenantId() || 'default'
    return navigator.locks.request(`nimbus-auth-token-refresh-${tenantId}`, executeRefresh)
  }
  return executeRefresh()
}

const getRefreshedAccessToken = (expiredAccessToken?: string): Promise<string> => {
  if (!refreshPromise) {
    refreshPromise = refreshAccessTokenWithLock(expiredAccessToken).finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

const handleUnauthorized = async (config: AuthRequestConfig) => {
  if (config.__authRetry) {
    return expireSession()
  }
  config.__authRetry = true
  const expiredAccessToken = config.headers.Authorization?.toString().replace(/^Bearer\s+/, '')
  try {
    const refreshedAccessToken = await getRefreshedAccessToken(expiredAccessToken)
    config.headers.Authorization = 'Bearer ' + refreshedAccessToken
    if (config.headers.isEncrypt) {
      config.headers.isEncrypted = true
    }
    return service(config)
  } catch {
    return expireSession()
  }
}

const expireSession = (): Promise<never> => {
  const { t } = useI18n()
  if (!sessionExpiryStarted) {
    sessionExpiryStarted = true
    removeToken()
    deleteUserCache()
    resetRouter()
    ElMessage.closeAll()
    ElNotification.closeAll()

    if (router.currentRoute.value.path !== '/login') {
      const browserFullPath =
        window.location.pathname + window.location.search + window.location.hash
      const redirect =
        router.currentRoute.value.fullPath === '/' && browserFullPath !== '/'
          ? browserFullPath
          : router.currentRoute.value.fullPath
      ElMessage.warning(t('sys.api.timeoutMessage'))
      const loginLocation = router.resolve({ path: '/login', query: { redirect } })
      window.location.replace(loginLocation.href)
    }
  }
  return Promise.reject(new Error('SESSION_EXPIRED'))
}
export { service }
