<template>
  <main :class="prefixCls" class="nimbus-login">
    <section class="nimbus-login__story" aria-labelledby="nimbus-login-title">
      <div class="nimbus-login__brand" translate="no">
        <img src="@/assets/svgs/nimbus-mark.svg" alt="Nimbus Framework" width="48" height="48" />
        <strong>Nimbus<br />Framework</strong>
      </div>

      <div class="nimbus-login__statement">
        <h1 id="nimbus-login-title">统一管理 · 高效运营 · 安全可控</h1>
        <p>
          Nimbus Framework 帮助团队构建统一的运营与管理平台，<br />
          实现多租户隔离、精细化权限控制与全链路审计。
        </p>
        <div class="nimbus-login__visual" aria-hidden="true">
          <span class="orbit orbit--outer"></span>
          <span class="orbit orbit--inner"></span>
          <span class="cube cube--main"></span>
          <span class="cube cube--left"></span>
          <span class="cube cube--right"></span>
        </div>
      </div>
    </section>

    <section class="nimbus-login__panel" aria-label="登录 Nimbus Framework">
      <div class="nimbus-login__toolbar">
        <ThemeSwitch />
        <LocaleDropdown />
      </div>
      <div class="nimbus-login__form-shell">
        <div class="nimbus-login__form-brand" translate="no">
          <img src="@/assets/svgs/nimbus-mark.svg" alt="Nimbus Framework" width="48" height="48" />
          <strong>Nimbus<br />Framework</strong>
        </div>
        <div class="nimbus-login__form-intro">
          <h2>欢迎登录</h2>
          <p>统一的企业运营与管理平台</p>
        </div>
        <LoginForm />
        <MobileForm />
        <QrCodeForm />
        <RegisterForm />
        <SSOLoginVue />
        <ForgetPasswordForm />
      </div>
    </section>

    <footer class="nimbus-login__footer">
      <span>© 2026 Nimbus Framework. 保留所有权利。</span>
      <nav aria-label="登录页辅助链接">
        <span>版本 v1.2.0</span>
        <span>隐私政策</span>
        <span>使用条款</span>
        <span>帮助中心</span>
      </nav>
    </footer>
  </main>
</template>

<script lang="ts" setup>
import { useDesign } from '@/hooks/web/useDesign'
import { ThemeSwitch } from '@/layout/components/ThemeSwitch'
import { LocaleDropdown } from '@/layout/components/LocaleDropdown'
import {
  LoginForm,
  MobileForm,
  QrCodeForm,
  RegisterForm,
  SSOLoginVue,
  ForgetPasswordForm
} from './components'

defineOptions({ name: 'Login' })

const { getPrefixCls } = useDesign()
const prefixCls = getPrefixCls('login')
</script>

<style lang="scss" scoped>
.nimbus-login {
  display: grid;
  min-height: 100vh;
  padding-bottom: 64px;
  overflow: hidden auto;
  color: var(--text-primary);
  background:
    radial-gradient(circle at 13% 14%, rgb(60 99 243 / 8%), transparent 28%), var(--bg-canvas);
  grid-template-columns: minmax(0, 55fr) minmax(440px, 45fr);

  &__story {
    display: flex;
    min-width: 0;
    min-height: calc(100vh - 64px);
    padding: var(--space-10) var(--space-12);
    flex-direction: column;
  }

  &__brand,
  &__form-brand {
    display: flex;
    gap: var(--space-3);
    align-items: center;

    img {
      flex: none;
    }

    strong {
      font-size: 18px;
      line-height: 22px;
      letter-spacing: -0.02em;
    }
  }

  &__statement {
    display: flex;
    max-width: 720px;
    padding: var(--space-12) var(--space-10) 0;
    margin: auto 0;
    flex-direction: column;

    h1 {
      margin: 0 0 var(--space-5);
      font-size: clamp(30px, 2.6vw, 42px);
      font-weight: 600;
      line-height: 1.28;
      letter-spacing: 0.02em;
      text-wrap: balance;
      overflow-wrap: anywhere;
    }

    p {
      font-size: 16px;
      line-height: 2;
      color: var(--text-secondary);
      text-wrap: pretty;
    }
  }

  &__visual {
    position: relative;
    width: min(620px, 100%);
    height: 260px;
    margin-top: var(--space-8);
    overflow: hidden;
    background: radial-gradient(ellipse at center, rgb(60 99 243 / 9%), transparent 58%);
  }

  .orbit {
    position: absolute;
    top: 58%;
    left: 50%;
    border: 1px solid rgb(60 99 243 / 13%);
    border-radius: 50%;
    transform: translate(-50%, -50%) rotate(-10deg);

    &--outer {
      width: 520px;
      height: 170px;
    }

    &--inner {
      width: 310px;
      height: 105px;
    }
  }

  .cube {
    position: absolute;
    display: block;
    width: 34px;
    height: 34px;
    background: linear-gradient(145deg, rgb(255 255 255 / 92%), rgb(60 99 243 / 25%));
    border: 1px solid rgb(60 99 243 / 18%);
    border-radius: var(--radius-md);
    transform: rotate(30deg) skew(-4deg);
    box-shadow: 0 12px 28px rgb(60 99 243 / 12%);

    &--main {
      top: 88px;
      left: calc(50% - 45px);
      width: 90px;
      height: 90px;
    }

    &--left {
      top: 148px;
      left: 26%;
    }

    &--right {
      top: 70px;
      right: 24%;
      width: 28px;
      height: 28px;
    }
  }

  &__panel {
    position: relative;
    display: flex;
    min-width: 0;
    min-height: calc(100vh - 64px);
    padding: var(--space-10) var(--space-12);
    align-items: center;
    justify-content: center;
  }

  &__toolbar {
    position: absolute;
    top: var(--space-6);
    right: var(--space-8);
    display: flex;
    gap: var(--space-2);
  }

  &__form-shell {
    width: min(100%, 520px);
    padding: var(--space-16) var(--space-12);
    background: var(--bg-surface);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-xl);
  }

  &__form-brand {
    margin-bottom: var(--space-8);
  }

  &__form-intro {
    margin-bottom: var(--space-6);

    h2 {
      margin: 0 0 var(--space-2);
      font-size: 28px;
      font-weight: 600;
      line-height: 36px;
      letter-spacing: -0.02em;
    }

    p {
      font-size: 14px;
      color: var(--text-tertiary);
    }
  }

  &__footer {
    position: absolute;
    bottom: 0;
    left: 0;
    display: flex;
    width: 100%;
    height: 64px;
    padding: 0 var(--space-12);
    font-size: 12px;
    color: var(--text-tertiary);
    background: rgb(255 255 255 / 42%);
    border-top: 1px solid var(--divider);
    align-items: center;
    justify-content: space-between;

    nav {
      display: flex;
      gap: var(--space-8);
    }
  }
}

:deep(.login-form) {
  width: 100%;

  > .el-row > .el-col:first-child {
    display: none;
  }

  .el-form-item__label {
    padding-bottom: var(--space-2);
    font-weight: 500;
    color: var(--text-primary);
  }

  .el-input__wrapper {
    min-height: 44px;
    background: var(--bg-surface);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-md);
    box-shadow: none;
  }

  .el-input__wrapper:hover {
    border-color: var(--border-strong);
  }

  .el-input__wrapper.is-focus {
    border-color: var(--primary);
    box-shadow: 0 0 0 3px var(--focus-ring);
  }

  .el-button--primary {
    min-height: 44px;
    font-weight: 500;
    background: var(--primary);
    border-color: var(--primary);
    border-radius: var(--radius-md);
  }

  .el-button--primary:hover {
    background: var(--primary-hover);
    border-color: var(--primary-hover);
  }
}

.dark .nimbus-login__footer {
  background: rgb(17 26 42 / 72%);
}

@media (width <= 1024px) {
  .nimbus-login {
    display: block;
    padding-bottom: 0;

    &__story,
    &__footer {
      display: none;
    }

    &__panel {
      min-height: 100vh;
      padding: 76px var(--space-6) var(--space-8);
    }

    &__form-shell {
      padding: var(--space-10) var(--space-8);
    }
  }
}

@media (width <= 560px) {
  .nimbus-login {
    &__panel {
      padding-right: var(--space-4);
      padding-left: var(--space-4);
    }

    &__toolbar {
      right: var(--space-4);
    }

    &__form-shell {
      padding: var(--space-8) var(--space-5);
    }
  }
}
</style>
