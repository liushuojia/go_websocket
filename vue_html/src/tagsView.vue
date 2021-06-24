<template>
  <div class="tags-view-wrapper">
    <router-link
      :to="item"
      :key="item.path"
      v-for="item in visitedTags"
      style="margin-right: 10px"
    >
      {{ item.name }}
      <span @click.prevent.stop="closeSelectedTag(item)"></span>
    </router-link>
  </div>
</template>
<script>
import { useStore } from 'vuex'
export default {
  setup() {
    const store = useStore()
    console.log(store.state.visitedTags)
    return {
      //必须返回 模板中才能使用
      visitedTags: store.state.visitedTags,
      name: 'abc',
    }
  },
}
</script>
