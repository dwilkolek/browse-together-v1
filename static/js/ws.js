function createQuerySelector(element) {
  if (element.id) {
    return `#${element.id}`;
  }

  const path = [];
  let currentElement = element;
  let error = false;

  while (currentElement.tagName !== "BODY") {
    const parent = currentElement.parentElement;

    if (!parent) {
      error = true;
      break;
    }

    const childTagCount = {};
    let nthChildFound = false;

    for (const child of parent.children) {
      const tag = child.tagName;
      const count = childTagCount[tag] || 0;
      childTagCount[tag] = count + 1;

      if (child === currentElement) {
        nthChildFound = true;
        break;
      }
    }

    if (!nthChildFound) {
      error = true;
      break;
    }

    const count = childTagCount[currentElement.tagName];
    const tag = currentElement.tagName.toLowerCase();
    const selector = `${tag}:nth-of-type(${count})`;

    path.push(selector);
    currentElement = parent;
  }

  if (error) {
    // console.error("Unable to create query selector", element);
    return;
  }

  path.push("body");

  const querySelector = path.reverse().join(" > ");

  return querySelector;
}

const blankPosition = JSON.stringify({
  elementQuery: undefined,
  x: -1,
  y: -1,
});

let sock = null;
const referenceBoxes = ["section"];
let lastPosition = blankPosition;
addEventListener("pointermove", function (ev) {
  if (sock) {
    const hovered = document.querySelectorAll(":hover");
    let selectedEl = hovered[hovered.length - 1];
    let valid = false;
    for (let i = hovered.length - 1; i > 0; i--) {
      const el = hovered[i];
      if (referenceBoxes.includes(el.nodeName.toLocaleLowerCase())) {
        selectedEl = el;
        valid = true;
        break;
      }
    }
    // if (!valid && lastPosition == blankPosition) {
    //   return
    // }

    let nextPosition = blankPosition;
    if (valid) {
      const rect = selectedEl.getBoundingClientRect();
      const querySelector = createQuerySelector(selectedEl);

      if (querySelector) {
        nextPosition = JSON.stringify({
          elementQuery: createQuerySelector(selectedEl),
          x: (ev.clientX - rect.left) / rect.width,
          y: (ev.clientY - rect.top) / rect.height,
        });
      }
    }
    
    if (nextPosition != lastPosition) {
      lastPosition = nextPosition;
      sock.send(lastPosition);
    }
  }
});

async function createSession() {
  const response = await fetch("/api/v1/sessions", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      name: "aaa",
      baseLocation: "aaa",
      creator: "aaa",
    }),
  });
  const session = await response.json();
  sock = new WebSocket(
    `${window.location.origin.replace("http", "ws")}/ws/${session.id}/cursors`
  );
  sock.onmessage = (event) => {
    const positions = JSON.parse(event.data ?? "[]");
    let notHandled = [...positions];
    document.querySelectorAll(".cursor")?.forEach((c) => {
      newPosition = positions.find((p) => p.clientId == c.dataset.clientId);
      if (newPosition) {
        notHandled = notHandled.filter(
          (n) => n.clientId !== newPosition.clientId
        );
        updateNodePosition(newPosition, c);
        c.style.top = newPosition.top;
      } else {
        c.remove();
      }
    });
    for (const cursorPosition of notHandled) {
      const cursorEl = document.createElement("div");
      cursorEl.dataset.clientId = cursorPosition.clientId;

      cursorEl.classList.add("cursor");

      updateNodePosition(cursorPosition, cursorEl);
      document.body.appendChild(cursorEl);
    }
  };
}

function updateNodePosition(cursorPosition, element) {
  const boxElement = document.querySelector(cursorPosition.elementQuery);
  const boxElementRect = boxElement.getBoundingClientRect();

  const cursorEl = document.createElement("div");

  cursorEl.classList.add("cursor");
  element.style.left = `${Math.round(
    boxElementRect.width * cursorPosition.x +
      boxElementRect.left +
      window.scrollX
  )}px`;
  element.style.top = `${Math.round(
    boxElementRect.height * cursorPosition.y +
      boxElementRect.top +
      window.scrollY
  )}px`;
}
