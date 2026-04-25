/**
 * 控制台 API 端点
 * 提供 VNC/SPICE 远程控制台接入功能
 */
import { get, post } from './request'

/** VNC 代理票据响应 */
export interface VNCTicketResponse {
  /** VNC 连接端口 */
  port: number
  /** VNC 连接票据 */
  ticket: string
  /** TLS 证书（可选） */
  cert?: string
  /** PVE 任务 ID */
  upid: string
}

/** SPICE 代理响应 */
export interface SPICEProxyResponse {
  /** SPICE 代理 URL */
  proxy: string
  /** SPICE 连接票据 */
  ticket: string
}

/**
 * 获取 QEMU 虚拟机的 VNC 代理票据
 * 用于 noVNC WebSocket 连接前的身份验证
 *
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @returns VNC 票据信息
 */
export async function getQEMUVNCTicket(
  node: string,
  vmid: number,
): Promise<VNCTicketResponse> {
  return post<VNCTicketResponse>(
    `/pve/nodes/${node}/qemu/${vmid}/vncproxy`,
    { websocket: 1 },
  )
}

/**
 * 获取 LXC 容器的 VNC 代理票据
 *
 * @param node 节点名称
 * @param vmid 容器 ID
 * @returns VNC 票据信息
 */
export async function getLXCVNCTicket(
  node: string,
  vmid: number,
): Promise<VNCTicketResponse> {
  return post<VNCTicketResponse>(
    `/pve/nodes/${node}/lxc/${vmid}/vncproxy`,
    { websocket: 1 },
  )
}

/**
 * 获取 SPICE 代理信息
 * 用于 SPICE 客户端连接（可选功能）
 *
 * @param node 节点名称
 * @param vmid 虚拟机 ID
 * @returns SPICE 代理信息
 */
export async function getSPICEProxy(
  node: string,
  vmid: number,
): Promise<SPICEProxyResponse> {
  return post<SPICEProxyResponse>(
    `/pve/nodes/${node}/qemu/${vmid}/spiceproxy`,
    {},
  )
}
