import { expect, test, vi } from 'vitest'
import axios from 'axios'
import router from './router'
import { apiConfig } from './config/api'

test.skipIf(!process.env.FRONTEND_LIVE_TEST)('首页使用运行中的后端数据渲染新闻及时间', async () => {
  document.body.innerHTML = '<div id="app"></div>'
  window.scrollTo = vi.fn()
  const response = await axios.get(`${apiConfig.baseURL}/api/news/list`, { params: { categoryId: 1, pageSize: 10 } })
  expect(response.data.code).toBe(200)
  const item = response.data.data.list[0]
  expect(item).toBeDefined()
  expect(item.publishTime).toBeTruthy()
  try {
    await import('./main')
    await router.isReady()
    await vi.waitFor(() => {
      expect(document.querySelector('.news-title')?.textContent).toBe(item.title)
      expect(document.querySelector('.news-info')?.textContent).toContain(item.publishTime)
    }, { timeout: 5000 })
  } finally {
    document.querySelector('#app')?.__vue_app__?.unmount()
  }
})
