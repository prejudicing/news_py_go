import { beforeEach, expect, test, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import axios from 'axios'
import { useNewsStore } from './news'

beforeEach(() => {
  vi.restoreAllMocks()
  setActivePinia(createPinia())
})

const response = (id, hasMore = false) => ({ data: { code: 200, data: { list: [{ id }], hasMore } } })

test('重复加载只发送一次请求，并遵守 hasMore', async () => {
  let resolve
  const get = vi.spyOn(axios, 'get').mockReturnValue(new Promise(done => { resolve = done }))
  const store = useNewsStore()
  const first = store.getNewsList()
  await store.getNewsList()
  expect(get).toHaveBeenCalledTimes(1)
  resolve(response(1))
  await first
  await store.getNewsList()
  expect(get).toHaveBeenCalledTimes(1)
  expect(store.newsList).toEqual([{ id: 1 }])
  expect(store.finished).toBe(true)
})

test('分类切换时丢弃旧分类的延迟响应', async () => {
  const resolvers = []
  vi.spyOn(axios, 'get').mockImplementation(() => new Promise(done => resolvers.push(done)))
  const store = useNewsStore()
  const oldRequest = store.getNewsList()
  const newRequest = store.changeCategory(2)
  resolvers[1](response(2))
  await newRequest
  resolvers[0](response(1))
  await oldRequest
  expect(store.newsList).toEqual([{ id: 2 }])
  expect(store.currentCategory).toBe(2)
  expect(store.loading).toBe(false)
})

test('刷新时丢弃旧页响应，下一页从第二页继续', async () => {
  const resolvers = []
  const get = vi.spyOn(axios, 'get').mockImplementation(() => new Promise(done => resolvers.push(done)))
  const store = useNewsStore()
  const oldRequest = store.getNewsList()
  const refresh = store.getNewsList(true)
  resolvers[1](response(2, true))
  await refresh
  resolvers[0](response(1, true))
  await oldRequest
  const next = store.getNewsList()
  expect(get.mock.calls[2][1].params.page).toBe(2)
  resolvers[2](response(3))
  await next
  expect(store.newsList).toEqual([{ id: 2 }, { id: 3 }])
})
