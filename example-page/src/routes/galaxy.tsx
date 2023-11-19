import { ActionFunctionArgs, useLoaderData } from "react-router-dom";
import { Galaxy, Planet } from "./root";
import { universe } from "./universe-data";

export default function GalaxyPage() {
  const data = useLoaderData();

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const galaxy: Galaxy = (data as any).galaxy;
  return (
    <div className="galaxy">
      <h2>{galaxy.name}</h2>
      <div className="planets">
        {galaxy.planets.map((p) => (
          <PlanetComponent key={p.id} planet={p} />
        ))}
      </div>
    </div>
  );
}

function PlanetComponent({ planet }: { planet: Planet }) {
  return (
    <div
      className="planet"
      style={{
        width: planet.size,
        height: planet.size,
        borderRadius: planet.size,
      }}
    >
      {planet.name}
    </div>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export async function galaxyLoader({ params }: ActionFunctionArgs<unknown>) {
  return { galaxy: universe.find((g) => `${g.id}` == params.galaxyId)! };
}
