version 1

let points = input.base.responseTemplate.body.points
assert len(points), "base response must contain at least one point"

let closest = points[0]

emit "get-closest-store" {
  requestMatch {
    method: "GET"
    path: "/api/v1/get-closest-store"
  }
  responseTemplate {
    status: 200
    body: {
      store: closest
    }
  }
  meta {
    source: "points"
    count: len(points)
  }
}

