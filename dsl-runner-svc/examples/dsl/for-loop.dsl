version 1

for item, idx in input.params.items {
  emit item.name + "-" + idx {
    requestMatch {
      method: "GET"
      path: "/items/" + idx
    }
    responseTemplate {
      status: 200
      body: {
        id: item.id
        name: item.name
        index: idx
      }
    }
    meta {
      source: "params.items"
      i: idx
    }
  }
}

