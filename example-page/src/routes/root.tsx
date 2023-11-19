import { Link, Outlet } from "react-router-dom";
import { useEffect, useState } from "react";

import { BrowseTogetherSdk, Session } from "browse-together-sdk";
import { universe } from "./universe-data";

function createSdk(url: string): BrowseTogetherSdk {
  const isTrackedElement = (e: Element) => e.classList.contains("planet");

  const cursorFactory: (
    memberId: number,
    givenIdentifier: string
  ) => HTMLElement = (memberId: number, givenIdentifier: string) => {
    const cursor = document.createElement("div");
    cursor.innerText = `${memberId}|${givenIdentifier}`;
    cursor.style.background = "#bbaa44";
    cursor.style.padding = "4px";
    cursor.style.pointerEvents = "none";
    cursor.style.position = "absolute";
    return cursor;
  };
  const sdk = new BrowseTogetherSdk(url, isTrackedElement, cursorFactory);
  sdk.drawYourself = true;
  return sdk;
}

export default function Root() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [connectedTo, setConnectedTo] = useState<Session | null>(null);
  const [rejoinToken, setRejoinToken] = useState<string | null>(
    localStorage.getItem("rejoinToken")
  );
  const [rejoinSession, setRejoinSession] = useState<Session | null>(
    localStorage.getItem("rejoinSession")
      ? JSON.parse(localStorage.getItem("rejoinSession")!)
      : null
  );
  const [sdk, setSdk] = useState<BrowseTogetherSdk>(
    createSdk("http://localhost:8080")
  );
  useEffect(() => {
    sdk.onDisconnect = () => {
      setConnectedTo(null);
    };
  }, [sdk]);

  return (
    <>
      <div id="sidebar">
        <h1>Browse together the universe</h1>
        {rejoinToken} | {JSON.stringify(rejoinSession)}
        <div>
          {sdk && (
            <button
              onClick={async () => {
                setSessions(await sdk.getActiveSessions());
              }}
            >
              Fetch sessions
            </button>
          )}
          {sdk && (
            <button
              onClick={() =>
                sdk.createSession(`S-${new Date()}`, "planetarian")
              }
            >
              Create session
            </button>
          )}
          <button
            onClick={async () => {
              await sdk?.leaveSession();
              setSessions([]);
              setConnectedTo(null);
              const url =
                prompt("Browse Together URL", "http://localhost:8080") ??
                "http://localhost:8080";
              setSdk(createSdk(url));
            }}
          >
            Connect to
          </button>
        </div>
        {!connectedTo &&
          sessions.map((session) => (
            <div key={session.id}>
              {session.name}
              <button
                onClick={async () => {
                  const newRejoinToken = await sdk.joinSession(
                    session,
                    "Johhny Oil"
                  );
                  setConnectedTo(session);
                  setRejoinSession(session);
                  localStorage.setItem(
                    "rejoinSession",
                    JSON.stringify(session)
                  );
                  if (newRejoinToken) {
                    setRejoinToken(newRejoinToken);
                    localStorage.setItem("rejoinToken", newRejoinToken);
                  }
                }}
              >
                Join!
              </button>
            </div>
          ))}
        {!connectedTo && rejoinToken && rejoinSession && (
          <button
            onClick={async () => {
              const newRejoinToken = await sdk.joinSession(
                rejoinSession,
                "Johnny Reoil",
                rejoinToken
              );
              if (newRejoinToken) {
                setRejoinToken(newRejoinToken);
                localStorage.setItem("rejoinToken", newRejoinToken);
              }
            }}
          >
            Rejoin {rejoinSession.name}!
          </button>
        )}
        <nav>
          <ul>
            {universe.map((g) => (
              <li key={g.id}>
                <a href={`/galaxy/${g.id}`}>{g.name} (reload)</a>
                <br />
                <Link to={`/galaxy/${g.id}`}>{g.name} (noreload)</Link>
              </li>
            ))}
          </ul>
        </nav>
      </div>
      <div id="detail">
        <div className="planet">
          {connectedTo && (
            <>
              <button
                onClick={() => {
                  sdk.closeSession();
                }}
              >
                Close session
              </button>{" "}
              <br />
              <button
                onClick={() => {
                  sdk.leaveSession();
                }}
              >
                Leave session
              </button>
              <br />
            </>
          )}
        </div>
        <Outlet />
      </div>
    </>
  );
}

export type Planet = {
  id: number;
  size: number;
  name: string;
};
export type Galaxy = {
  id: number;
  name: string;
  planets: Planet[];
};
