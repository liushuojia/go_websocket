export default {
  obj: null,
  status: false,
  timeOutObj: null,
  errorNum: 0,
  leave: false,
  wsuri: undefined,
  init(wsuri, callbackFuc) {
    if (this.status) return
    if (this.leave) return

    //初始化weosocket
    if (!this.wsuri) {
      this.wsuri = wsuri
    }
    this.obj = new WebSocket(wsuri)
    this.obj.onmessage = (e) => {
      if (!this.status) return false

      try {
        const redata = JSON.parse(e.data)
        switch (redata.action) {
          case "heartbeat":
            this.send(
              JSON.stringify({
                action: "heartbeat",
              })
            )
            break
          default:
            callbackFuc(redata)
            break
        }
      } catch (err) {
        console.log(err)
      }
    }
    this.obj.onopen = () => {
      clearTimeout(this.timeOutObj)
      this.errorNum = 1
      this.status = true
      let actions = { message: "hello world" }
      this.send(JSON.stringify(actions))
    }
    this.obj.onerror = () => {
      console.log("出现错误")
      this.obj.close()
    }
    this.obj.onclose = (e) => {
      console.log("断开连接", e)
      this.status = false
      this.reconnect(callbackFuc)
      callbackFuc({
        action: "close",
      })
    }
  },
  send(Data) {
    if (!this.obj) return false
    if (!this.status) return false
    try {
      this.obj.send(Data)
      return true
    } catch (e) {
      return false
    }
    //数据发送
  },
  close() {
    if (!this.obj) return
    if (!this.status) return
    this.obj.close()
    return
  },
  reconnect(callbackFuc) {
    clearTimeout(this.timeOutObj)
    if (this.status) return
    if (this.leave) return

    this.errorNum = this.errorNum + 1
    let waitTime = this.errorNum * 1000 + 8000
    console.log("wait " + waitTime)
    this.timeOutObj = setTimeout(() => {
      if (this.status) return
      if (this.leave) return
      this.init(this.wsuri, callbackFuc)
    }, waitTime)
  },
}
