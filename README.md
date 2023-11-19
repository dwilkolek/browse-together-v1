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
function createSdk(url: string): BrowseTogetherSdk {
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

  const sdk = new BrowseTogetherSdk(url, isTrackedElement, cursorFactory);

  // if you want to see your cursor
  sdk.drawYourself = true;

  return sdk;
}
```



## Deploy backend to fly.dev

`make fly`
````
