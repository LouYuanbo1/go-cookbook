import type { Method,RawAxiosRequestHeaders ,ResponseType} from 'axios';

interface RequestConfig {
  url: string;
  method: Method;
  data?: any;
  params?: Record<string, any>;
  headers?: RawAxiosRequestHeaders;
  responseType?: ResponseType;
  rawResponse?: boolean;
}

import api from './interceptor'
/*
import axios from 'axios'

async function request(config: RequestConfig) {
  try {
    const resp = await axios({
      url: config.url,
      method: config.method,
      data: config.data,
      params: config.params,
      headers: config.headers,
    })
    console.log('resp:', resp)
    if (resp.status == 401) {
      return Promise.reject("非授权用户")
    }
    return resp.data
  } catch (err) {
    return Promise.reject(err)
  }
}

export default request
*/

async function request(config: RequestConfig) {
  try {
    const resp = await api({
      url: config.url,
      method: config.method,
      data: config.data,
      params: config.params,
      headers: config.headers,  // 自定义 headers 会与拦截器设置的合并
      responseType: config.responseType,
    })
    console.log('resp:', resp)
    // 注意：axios 默认只在状态码 2xx 时 resolve，非 2xx 会抛出错误，
    // 因此 resp.status == 401 的判断通常不会执行。如果确实需要自定义，可以配置 validateStatus
    if (resp.status === 401) {
      return Promise.reject("非授权用户")
    }
    if (config.rawResponse || config.responseType === 'blob' || config.responseType === 'stream') {
      return resp;
    }
    return resp.data
  } catch (err) {
    return Promise.reject(err)
  }
}

export default request