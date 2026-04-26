import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

/**
 * Axios 实例封装
 * - 统一设置 baseURL、超时时间等默认配置
 * - 请求拦截器：注入认证 Token
 * - 响应拦截器：统一处理错误码、401 自动跳转登录
 */

const service: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

/**
 * 请求拦截器
 * 从 localStorage 获取 token 并添加到请求头
 */
service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('pve_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

/**
 * 响应拦截器
 * - 统一处理业务错误码
 * - 401 未授权时清除 token 并跳转登录
 * - 其他错误显示提示信息
 */
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data
    // 假设后端统一响应格式: { code, data, message }
    if (res.code !== undefined && res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    // 成功时返回 data 字段
    return res.data !== undefined ? res.data : res
  },
  (error) => {
    const { response } = error
    // 404 错误静默返回，由调用方决定如何处理
    // 这样可以支持后端部分接口尚未实现时的优雅降级
    if (response?.status === 404) {
      return Promise.reject(error)
    }
    if (response) {
      switch (response.status) {
        case 401:
          // Token 过期或无效，清除本地状态并跳转登录
          localStorage.removeItem('pve_token')
          router.push('/login')
          ElMessage.error('登录已过期，请重新登录')
          break
        case 403:
          ElMessage.error('没有权限访问该资源')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          ElMessage.error(response.data?.message || '网络错误')
      }
    } else {
      ElMessage.error('网络连接异常')
    }
    return Promise.reject(error)
  },
)

/**
 * 封装的 GET 请求
 */
export function get<T = unknown>(url: string, params?: Record<string, unknown>, config?: AxiosRequestConfig): Promise<T> {
  return service.get(url, { params, ...config }) as Promise<T>
}

/**
 * 封装的 POST 请求
 */
export function post<T = unknown>(url: string, data?: Record<string, unknown>, config?: AxiosRequestConfig): Promise<T> {
  return service.post(url, data, config) as Promise<T>
}

/**
 * 封装的 PUT 请求
 */
export function put<T = unknown>(url: string, data?: Record<string, unknown>, config?: AxiosRequestConfig): Promise<T> {
  return service.put(url, data, config) as Promise<T>
}

/**
 * 封装的 DELETE 请求
 */
export function del<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return service.delete(url, config) as Promise<T>
}

export default service
