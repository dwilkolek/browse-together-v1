import { Galaxy } from "./root";

export const universe: Galaxy[] = [
    {
      id: 1,
      name: "Milkyway",
      planets: [
        {
          id: 1,
          size: 150,
          name: "Earth",
        },
        {
          id: 2,
          size: 300,
          name: "Mars",
        },
      ],
    },
    {
      id: 2,
      name: "Andromeda",
      planets: [
        {
          id: 1,
          size: 300,
          name: "PA-99-N2",
        },
        {
          id: 2,
          size: 200,
          name: "PA-99-N3",
        },
        {
          id: 3,
          size: 100,
          name: "PA-99-N4",
        },
        {
          id: 4,
          size: 450,
          name: "PA-99-N5",
        },
      ],
    },
  ];
  