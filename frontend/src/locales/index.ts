import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'

/**
 * 国际化配置
 * 默认使用中文，支持后续扩展英文等其他语言
 */
const i18n = createI18n({
  legacy: false,
  locale: 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
  },
})

export default i18n
