<template>
  <div>
    <div id="nav">
      <router-link to="/"> Home </router-link> |
      <router-link to="/about"> About </router-link>
    </div>

    <div>
      <tagsView />
      <!-- <keep-alive>
      <router-view v-if="keepAlive"></router-view>
    </keep-alive>
    <router-view v-if="!keepAlive"></router-view> -->
    </div>

    <router-view v-slot="{ Component }">
      <keep-alive v-if="keepAlive">
        <component :is="Component">
          <p>缓存页面</p>
        </component>
      </keep-alive>
      <component :is="Component" v-if="!keepAlive">
        <p>没有缓存的页面</p>
      </component>
    </router-view>
  </div>
</template>

<script>
import { watch, computed } from "vue"
import { useRoute, useRouter } from "vue-router"
import { useStore } from "vuex"
import tagsView from "./tagsView.vue"

export default {
  components: {
    tagsView,
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const store = useStore()
    console.log(route)
    console.log(router)
    function pageInit() {
      return
    }
    function addTags() {
      store.commit({
        type: "addTags",
        route,
      })
    }
    watch(route, () => {
      addTags()
    })
    let keepAlive = computed(() => {
      return true
      // return route.meta && route.meta.keepAlive ? true : false
    })
    return {
      //必须返回 模板中才能使用
      pageInit,
      addTags,
      keepAlive,
    }
  },
}
</script>
<style lang="less">
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

#nav {
  padding: 30px;

  a {
    font-weight: bold;
    color: #2c3e50;

    &.router-link-exact-active {
      color: #42b983;
    }
  }
}
</style>
