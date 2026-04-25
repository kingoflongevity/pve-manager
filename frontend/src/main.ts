import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

import App from './App.vue'
import router from './router'
import i18n from './locales'

import './assets/styles/global.scss'

/**
 * 应用初始化入口
 * 注册 Pinia 状态管理、Vue Router、Element Plus UI 库、国际化等插件
 */
const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })
app.use(i18n)

app.mount('#app')
