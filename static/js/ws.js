function createQuerySelector(element) {
  if (element.id) {
    return `#${element.id}`;
  }

  const path = [];
  let currentElement = element;
  let error = false;
  while (currentElement.tagName.toLocaleLowerCase() !== "body") {
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
let toSend = blankPosition;
let lastSend = blankPosition;
setInterval(() => {
  const msg = toSend;
  if (sock && toSend != lastSend) {
    console.log("Sending", blankPosition == msg);
    lastSend = toSend;
    sock.send(msg);
    return;
  } else {
    console.log("NOOP");
  }
}, 16);

addEventListener("pointermove", function (ev) {
  if (sock) {
    const hovered = document.querySelectorAll(":hover:not(.cursor)");
    let selectedEl;
    for (let i = hovered.length - 1; i > 0; i--) {
      const tmpEl = hovered[i];
      if (referenceBoxes.includes(tmpEl.nodeName.toLocaleLowerCase())) {
        selectedEl = tmpEl;
        break;
      }
    }

    let nextPosition = blankPosition;
    if (selectedEl) {
      const rect = selectedEl.getBoundingClientRect();
      const querySelector = createQuerySelector(selectedEl);

      if (querySelector) {
        nextPosition = JSON.stringify({
          elementQuery: querySelector,
          x: (ev.clientX - rect.left) / rect.width,
          y: (ev.clientY - rect.top) / rect.height,
        });
      }
    }
    toSend = nextPosition;
  }
});

async function createSession() {
  const session = await apiCreateSession();
  joinSession(session);
}
const cursors = [];
function joinSession(session) {
  sock = new WebSocket(
    `${window.location.origin.replace("http", "ws")}/ws/${session.id}/cursors`
  );
  sock.onmessage = (event) => {
    const positions = JSON.parse(event.data ?? "[]");
    for (const cursorPosition of positions) {
      let cursorEl = cursors[cursorPosition.clientId];
      if (!cursorEl) {
        const cursorEl = document.createElement("div");

        cursorEl.dataset.clientId = cursorPosition.clientId;
        cursorEl.classList.add("cursor", `cursor-${cursorPosition.clientId}`);
        updateNodePosition(cursorPosition, cursorEl);
        document.body.appendChild(cursorEl);
        cursors[cursorPosition.clientId] = cursorEl;
      } else {
        updateNodePosition(cursorPosition, cursorEl);
      }
    }
  };
}

function updateNodePosition(cursorPosition, element) {
  const boxElement = document.querySelector(cursorPosition.elementQuery);
  const boxElementRect = boxElement.getBoundingClientRect();
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
