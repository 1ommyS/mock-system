version 1

assert input.params.region, "params.region is required"

let region = lower(input.params.region)
let includeDebug = get(input.params.debug, false)

if includeDebug {
  emit "debug-" + region {
    requestMatch {
      method: "GET"
      path: "/debug/" + region
    }
    responseTemplate {
      status: 200
      body: {
        mode: "debug"
        region: region
        hash: sha256(region)
      }
    }
    meta {
      branch: "debug"
    }
  }
} else {
  emit "prod-" + region {
    requestMatch {
      method: "GET"
      path: "/prod/" + region
    }
    responseTemplate {
      status: 200
      body: {
        mode: "prod"
        region: region
      }
    }
    meta {
      branch: "prod"
    }
  }
}

