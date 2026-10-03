import { afterAll, beforeAll, expect, test, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import axios from 'axios'
import router from './router'
import { useUserStore } from './store/user'
import { aiChatConfig } from './config/api'

const news = {
  id: 1, title: '前端验证新闻', description: '接口数据测试', author: '测试作者',
  categoryId: 1, views: 2, publishTime: '2026-10-03T12:00:00',
}

beforeAll(async () => {
  document.body.innerHTML = '<div id="app"></div>'
  window.scrollTo = vi.fn()
  vi.spyOn(axios, 'get').mockImplementation(async (url) => {
    if (url.includes('/api/news/categories')) {
      return { data: { code: 200, data: [{ id: 1, name: '头条' }] } }
    }
    if (url.includes('/api/news/list')) {
      return { data: { code: 200, data: { list: [news], total: 1, hasMore: false } } }
    }
    if (url.includes('/api/news/detail')) {
      return { data: { code: 200, data: { ...news, content: '新闻正文', relatedNews: [] } } }
    }
    throw new Error(`Unexpected test request: ${url}`)
  })
  await import('./main')
  await router.isReady()
  await flushPromises()
  await vi.waitFor(() => expect(document.querySelector('.news-title')?.textContent).toBe(news.title))
})

afterAll(() => {
  document.querySelector('#app')?.__vue_app__?.unmount()
  vi.restoreAllMocks()
})

test('首页渲染新闻和接口提供的发布时间', () => {
  expect(document.querySelector('.news-info').textContent).toContain(news.publishTime)
  expect(document.querySelector('.news-info').textContent).toContain(news.author)
})

test('新闻详情路由渲染正文', async () => {
  await router.push('/news/detail/1')
  await flushPromises()
  await vi.waitFor(() => expect(document.querySelector('.detail-content')?.textContent).toContain('新闻正文'))
})

test('AI页面未登录时不发送请求', async () => {
  const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue({})
  await router.push('/aichat')
  await flushPromises()
  const input = document.querySelector('textarea')
  input.value = '你好'
  input.dispatchEvent(new Event('input', { bubbles: true }))
  await flushPromises()
  document.querySelector('.send-button').click()
  await flushPromises()
  expect(fetchMock).not.toHaveBeenCalled()
  expect(document.body.textContent).toContain('请先登录后使用AI问答')
  fetchMock.mockRestore()
})

test('AI页面发送项目令牌并显示流式回复', async () => {
  useUserStore().$patch({ isLogin: true, token: 'local-session-test-token' })
  const bytes = new TextEncoder().encode('data: {"choices":[{"delta":{"content":"测试回复"}}]}\n\ndata: [DONE]\n\n')
  let consumed = false
  const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue({
    ok: true,
    body: { getReader: () => ({ read: async () => {
      if (consumed) return { done: true }
      consumed = true
      return { done: false, value: bytes }
    } }) },
  })
  document.querySelector('.send-button').click()
  await vi.waitFor(() => expect(document.querySelector('.messages-container').textContent).toContain('测试回复'))
  const [url, options] = fetchMock.mock.calls[0]
  expect(url).toBe(aiChatConfig.apiEndpoint)
  expect(url).toContain('/api/ai/chat')
  expect(options.headers.Authorization).toBe('local-session-test-token')
  expect(JSON.parse(options.body)).toEqual({ messages: expect.any(Array) })
  expect(aiChatConfig).not.toHaveProperty('apiKey')
  fetchMock.mockRestore()
})
