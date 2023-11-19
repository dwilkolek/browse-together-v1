# Browse Together V1

## Development

1. Run backend

   - `make dev` to run in memory
   - `make dev_redis` to run on redis

1. Build SDK.

   - `cd sdk && npm ci && npm run build`

1. Run example page
   
   - `cd example-page && npm ci && npm run dev`

## How to integrate

```
// backendUrl = http://localhost:8080/
function createSdk(backendUrl: string): BrowseTogetherSdk {
  const isTrackedElement = (e: Element) => e.classList.contains("planet");

  // cursorFactory is optional
  const cursorFactory: (
    memberId: number,
    givenIdentifier: string
  ) => HTMLElement = (memberId: number, givenIdentifier: string) => {
    const cursor = document.createElement("div");
    cursor.innerText = `${memberId}|${givenIdentifier}`;
    cursor.style.background = "#bbaa44";
    cursor.style.padding = "4px";
    return cursor
  };

  const sdk = new BrowseTogetherSdk(backendUrl, isTrackedElement, cursorFactory);

  // if you want to see your cursor
  sdk.drawYourself = true;

  return sdk;
}
```

## Deploy backend to fly.dev

`make fly`

## Problems:

Handling sessions with that architecture is pretty hard cause backend needs to keep state of all cursors. In worst case scenario(and very likely with round robin load balancing) all backend machines will keep state of all sessions and cursors. 
Using JSON is not a greatest idea. There is limited set of messages, using semicolon separated string would do the justice and would be easier than unmarshaling json.
