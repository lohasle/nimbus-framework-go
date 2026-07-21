/**
 * 配置浏览器本地存储的方式，可直接存储对象数组。
 */

import WebStorageCache from 'web-storage-cache'

type CacheType = 'localStorage' | 'sessionStorage'

const CACHE_NAMESPACE = 'nimbus-go:'

export const CACHE_KEY = {
  // 用户相关
  ROLE_ROUTERS: `${CACHE_NAMESPACE}roleRouters`,
  USER: `${CACHE_NAMESPACE}user`,
  VisitTenantId: `${CACHE_NAMESPACE}visitTenantId`,
  // 系统设置
  IS_DARK: `${CACHE_NAMESPACE}isDark`,
  UI_THEME_VERSION: `${CACHE_NAMESPACE}uiThemeVersion`,
  LANG: `${CACHE_NAMESPACE}lang`,
  THEME: `${CACHE_NAMESPACE}theme`,
  LAYOUT: `${CACHE_NAMESPACE}layout`,
  DICT_CACHE: `${CACHE_NAMESPACE}dictCache`,
  // 登录表单
  LoginForm: `${CACHE_NAMESPACE}loginForm`,
  TenantId: `${CACHE_NAMESPACE}tenantId`
}

export const useCache = (type: CacheType = 'localStorage') => {
  const wsCache: WebStorageCache = new WebStorageCache({
    storage: type
  })

  return {
    wsCache
  }
}

export const deleteUserCache = () => {
  const { wsCache } = useCache()
  wsCache.delete(CACHE_KEY.USER)
  wsCache.delete(CACHE_KEY.ROLE_ROUTERS)
  wsCache.delete(CACHE_KEY.VisitTenantId)
  // 注意，不要清理 LoginForm 登录表单
}
