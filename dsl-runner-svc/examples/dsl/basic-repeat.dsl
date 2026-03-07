version 1

let basePath = input.base.request_match.path

repeat 2 {
  emit "gen-" + loop.index {
    requestMatch {
      method: input.base.request_match.method
      path: basePath + "/v" + loop.index
    }
    responseTemplate {
      status: 200
      body: {
        stableId: randUUID()
        n: randInt(1, 10)
      }
    }
    meta {
      from: "dsl"
      i: loop.index
    }
  }
}

