export type SessionMember = {
  givenIdentifier: string;
  memberId: number;
  x: number;
  y: number;
  selector: string | undefined;
  location: string | undefined;
};

export type Session = {
  id: string;
  joinUrl: string;
  name: string;
  base: string;
  creatorIdentifier: string;
};

type UpdatePositionCmd = {
  x: number;
  y: number;
  selector: string | undefined;
  location: string | undefined;
};

type ValidLocator = {
  x: number;
  y: number;
  selector: string;
};

const OUT_OF_TRACKING_MESSAGE: UpdatePositionCmd = {
  x: -1,
  y: -1,
  selector: undefined,
  location: undefined,
};

const OUT_OF_TRACKING_MESSAGE_JSON = JSON.stringify(OUT_OF_TRACKING_MESSAGE);

export class BrowseTogetherSdk {
  private cursors: HTMLElement[] = [];
  private sock: WebSocket | undefined = undefined;
  private toSend = OUT_OF_TRACKING_MESSAGE_JSON;
  private lastSend = OUT_OF_TRACKING_MESSAGE_JSON;
  private sender: number | undefined;
  private session: Session | undefined = undefined;
  private memberId: number | undefined = undefined;
  private rejoinToken: string | undefined = undefined;
  constructor(
    private url: String,
    public isTrackedElement: (e: Element) => boolean = () => true,
    public cursorFactory: (
      memberId: number,
      givenIdentifier: string
    ) => HTMLElement = (memberId: number, givenIdentifier: string) => {
      const cursor = document.createElement("div");
      cursor.dataset.memberid = `${memberId}`;
      cursor.dataset.givenIdentifier = givenIdentifier;
      cursor.style.background = "#ad0c78";
      cursor.style.width = "12px";
      cursor.style.height = "12px";
      cursor.style.borderRadius = "50px";
      cursor.style.marginLeft = "-6px";
      cursor.style.marginTop = "-6px";
      return cursor;
    },
  ) {}

  public drawYourself = false;

  public onDisconnect: () => void = () => {};

  public get isConnected() {
    return !!this.session;
  }

  public async getActiveSessions(): Promise<Session[]> {
    const response = await fetch(`${this.url}/api/v1/sessions`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });
    return await response.json();
  }
  public async createSession(
    sessionName: string,
    creatorIdentifier: string
  ): Promise<Session> {
    const response = await fetch(`${this.url}/api/v1/sessions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name: sessionName,
        baseLocation: window.location.href,
        creator: creatorIdentifier,
      }),
    });
    return await response.json();
  }

  public joinSession(
    session: Session,
    givenIdentifier: string,
    rejoinToken: string | undefined = undefined
  ): Promise<string | undefined> {
    return new Promise<string | undefined>(async (resolve, reject) => {
      this.session = session;
      const response = await fetch(
        `${this.url}/api/v1/sessions/${session.id}/join`,
        {
          method: "POST",
        }
      );
      const json: { joinUrl: string } = await response.json();
      const joinUrl = json.joinUrl.startsWith("/")
        ? this.url.replace(/(http)(s)?\:\/\//, "ws$2://") +
          json.joinUrl +
          (rejoinToken ? "?rejoinToken=" + rejoinToken : "")
        : json.joinUrl;

      this.sock = new WebSocket(joinUrl);
      this.sock.onclose = () => {
        this.cleanup();
      };
      this.sock.onmessage = (event: MessageEvent<string>) => {
        if (this.memberId == null) {
          const data = JSON.parse(event.data).split(";");
          this.memberId = parseInt(data[0]);
          this.rejoinToken = data[1];
          resolve(this.rejoinToken);
          return;
        } else {
          reject("TODO");
        }
        const positions: SessionMember[] = JSON.parse(event.data ?? "[]");
        this.cursors.forEach((c) => {
          if (c) {
            c.style.display = "none";
          }
        });
        for (const member of positions) {
          let cursorEl = this.cursors[member.memberId];
          if (!member.selector || member.location !== window.location.href || (!this.drawYourself && member.memberId == this.memberId)) {
            continue;
          }
          const locator = member as ValidLocator;
          if (!cursorEl) {
            const cursorEl = this.cursorFactory(
              member.memberId,
              member.givenIdentifier
            );
            cursorEl.style.pointerEvents = "none"; 
            cursorEl.style.position = "absolute"; 

            this.updateNodePosition(locator, cursorEl);
            document.body.appendChild(cursorEl);
            this.cursors[member.memberId] = cursorEl;
          } else {
            this.updateNodePosition(locator, cursorEl);
          }
          this.cursors[member.memberId].style.display = "block";
        }
      };
      this.setup(givenIdentifier);
    });
  }

  public async leaveSession() {
    if (this.sock && this.isConnected) {
      this.sock.close();
      this.cleanup();
    }
  }

  public async closeSession() {
    if (this.session) {
      await fetch(`${this.url}/api/v1/sessions/${this.session.id}`, {
        method: "DELETE",
      });
    }
  }

  private createQuerySelector(element: Element): string {
    if (element.id) {
      return `#${element.id}`;
    }

    const path: string[] = [];
    let currentElement = element;
    let error = false;
    while (currentElement.tagName.toLocaleLowerCase() !== "body") {
      const parent = currentElement.parentElement;

      if (!parent) {
        error = true;
        break;
      }

      const childTagCount: { [index: string]: number } = {};
      let nthChildFound = false;

      for (
        let childIndex = 0;
        childIndex < parent.children.length;
        childIndex++
      ) {
        const child = parent.children.item(childIndex);
        if (!child) {
          break;
        }
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
      console.error("Unable to create query selector", element);
      throw Error("Unable to create query selector");
    }

    path.push("body");

    const querySelector = path.reverse().join(" > ");

    return querySelector;
  }

  private setup(identifier: string) {
    clearInterval(this.sender);
    this.sock?.addEventListener('open', () => {
      this.sock?.send("Identifier:" + identifier);
      this.sender = setInterval(() => {
        const msg = this.toSend;
        if (!this.isConnected) {
          clearInterval(this.sender);
          return;
        }
        if (this.sock && this.toSend != this.lastSend && this.isConnected) {
          this.lastSend = this.toSend;
          this.sock.send(msg);
        }
      }, 16);
  
      window.addEventListener("pointermove", (ev) => {
        if (this.sock && this.isConnected) {
          const hovered = document.querySelectorAll(":hover:not(.cursor)");
          let selectedEl: Element | undefined = undefined;
          for (let i = hovered.length - 1; i > 0; i--) {
            const tmpEl = hovered[i];
            if (this.isTrackedElement(tmpEl)) {
              selectedEl = tmpEl;
              break;
            }
          }
  
          let nextPosition = OUT_OF_TRACKING_MESSAGE_JSON;
          if (selectedEl) {
            const rect = selectedEl.getBoundingClientRect();
            const querySelector = this.createQuerySelector(selectedEl);
  
            if (querySelector) {
              const nextPositionUpdateCmd: UpdatePositionCmd = {
                location: window.location.href,
                selector: querySelector,
                x: (ev.clientX - rect.left) / rect.width,
                y: (ev.clientY - rect.top) / rect.height,
              };
              nextPosition = JSON.stringify(nextPositionUpdateCmd);
            }
          }
          this.toSend = nextPosition;
        }
      });
    })
   
    
  }

  private updateNodePosition(
    cursorPosition: ValidLocator,
    element: HTMLElement
  ) {
    const boxElement = document.querySelector(cursorPosition.selector);
    if (!boxElement) {
      throw Error("Element doesn't exist: " + cursorPosition.selector);
    }
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

  private cleanup() {
    this.memberId = undefined;
    this.rejoinToken = undefined;
    this.session = undefined;
    this.cursors.forEach((e) => e.remove());
    this.cursors = [];
    this.onDisconnect();
  }
}
