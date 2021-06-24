<template>
  <div>
    <div style="width: 100%">
      <a-form :label-col="{ span: 4 }" :wrapper-col="{ span: 18 }">
        <a-form-item label="host">
          <a-input v-model:value="connection.host" />
        </a-form-item>
        <a-form-item label="port">
          <a-input v-model:value="connection.port" />
        </a-form-item>
        <a-form-item label="endpoint">
          <a-input v-model:value="connection.endpoint" />
        </a-form-item>
        <a-form-item label="username">
          <a-input v-model:value="connection.username" />
        </a-form-item>
        <a-form-item label="password">
          <a-input v-model:value="connection.password" />
        </a-form-item>
        <a-form-item label="clientId">
          <a-input v-model:value="connection.clientId" />
        </a-form-item>
        <a-form-item :wrapper-col="{ span: 14, offset: 4 }">
          <a-space>
            <a-button @click="createConnection"> 连接 </a-button>
            <a-button @click="destroyConnection"> 断开 </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </div>

    <div style="margin-top: 60px">
      <a-form :label-col="{ span: 4 }" :wrapper-col="{ span: 18 }">
        <a-form-item label="topic">
          <a-input v-model:value="subscription.topic" />
        </a-form-item>
        <a-form-item :wrapper-col="{ span: 14, offset: 4 }">
          <a-space>
            <a-button @click="doSubscribe"> 订阅 </a-button>
            <a-button @click="doUnSubscribe"> 取消订阅 </a-button>
          </a-space>
        </a-form-item>

        <a-form-item label="message">
          {{ message }}
        </a-form-item>
      </a-form>

      { "topic":"go-mqtt/sample", "message": "Hello, I am browser." ,
      "clientId":"*" }
    </div>

    <div style="margin-top: 60px">
      <a-form :label-col="{ span: 4 }" :wrapper-col="{ span: 18 }">
        <a-form-item label="topic">
          <a-input v-model:value="publish.topic" />
        </a-form-item>

        <a-form-item label="payload">
          <a-input v-model:value="publish.payload" />
        </a-form-item>
        <a-form-item :wrapper-col="{ span: 14, offset: 4 }">
          <a-button @click="doPublish"> 发布消息 </a-button>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script>
import websocket from "@/utils/websocket"
export default {
  data() {
    return {
      connection: {
        host: "127.0.0.1",
        port: 8000,
        endpoint: "/mqtt",
        clean: false, // 保留会话
        connectTimeout: 4000, // 超时时间
        reconnectPeriod: 4000, // 重连时间间隔
        // 认证信息
        clientId: "mqttjs_3be2c321",
        username: "liushuojia",
        password: "password",
      },
      subscription: {
        topic: "go-mqtt/sample",
      },
      publish: {
        topic: "go-mqtt/sample",
        payload: '{ "message": "Hello, I am browser." }',
      },
      client: {
        connected: false,
      },
      receiveNews: "",
      qosList: [
        { label: 0, value: 0 },
        { label: 1, value: 1 },
        { label: 2, value: 2 },
      ],
      subscribeSuccess: false,
      message: undefined,
    }
  },
  created() {},
  methods: {
    createConnection() {
      // 连接字符串, 通过协议指定使用的连接方式
      // ws 未加密 WebSocket 连接
      // wss 加密 WebSocket 连接
      // mqtt 未加密 TCP 连接
      // mqtts 加密 TCP 连接
      // wxs 微信小程序连接
      // alis 支付宝小程序连接
      const { host, port, endpoint, username, password, clientId } =
        this.connection
      const connectUrl = `ws://${host}:${port}${endpoint}?username=${username}&password=${password}&clientId=${clientId}`

      if (websocket.status) {
        return
      }
      websocket.leave = false
      try {
        websocket.init(connectUrl, (data) => {
          this.message = data
        })
      } catch (error) {
        console.log("connect error", error)
      }
    },
    destroyConnection() {
      websocket.leave = true
      websocket.close()
    },

    doSubscribe() {
      const { topic } = this.subscription
      websocket.send(
        JSON.stringify({
          action: "subscribe",
          topic: topic,
        })
      )
    },
    doUnSubscribe() {
      const { topic } = this.subscription
      websocket.send(
        JSON.stringify({
          action: "unsubscribe",
          topic: topic,
        })
      )
    },
    doPublish() {
      const { topic, payload } = this.publish
      websocket.send(
        JSON.stringify({
          action: "publish",
          topic: topic,
          content: payload,
        })
      )
    },
  },
}
</script>

<style lang="less" type="text/less">
.ant-form-item {
  margin-bottom: 8px;
}
</style>
