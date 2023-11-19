# Browse together sdk

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