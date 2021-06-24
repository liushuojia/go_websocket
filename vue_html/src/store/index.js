import { createStore } from 'vuex'

export default createStore({
  state: {
    visitedTags: [],
  },
  mutations: {
    addTags(state, payload) {
      let flag = state.visitedTags.some(
        (item) => item.path === payload.route.path
      ) //打开标签后，判断数组中是否已经存在该路由
      if (!flag) {
        state.visitedTags.push({
          path: payload.route.path,
          name: payload.route.name,
          params: payload.route.params,
        })
      } //数组中路由存在不push ,单击左侧路由变化,点击标签路由变化均触发
    }, //添加标签
  },
  getters: {
    getTags: (state) => state.visitedTags,
  },
  actions: {},
  modules: {},
})
