import * as React from 'react';
import { SVGProps } from 'react';
import { useTheme, createTheme } from '@mui/material/styles';

const Logo = (props: SVGProps<SVGSVGElement>) => {
  const theme = useTheme();
  const logoTheme = createTheme({
    palette: {
      primary: {
        main: "#855E42"
      }
    },
  });
  return <svg
    id="e1690c7f-0f7f-4da9-91c6-93f6e68ff73e"
    data-name="Layer 1"
    xmlns="http://www.w3.org/2000/svg"
    width={217.532}
    height={36.475}
    viewBox="0 0 933.15 173.15"
    {...props}
  >
    <defs>
      <style>{`.bc8c4e7a-0a4c-47d4-8d3b-876939d268e6{fill:${theme.palette.secondary.light}}.a9fe37ec-4cf7-43eb-a57d-cd438e81ae89{fill:${theme.palette.primary.light}}`}</style>
    </defs>

    <g id="a1721df0-97d9-4765-85cf-aa11ad846227" data-name="Layer 2">
      <path
        d="M558.42 178.13a4.7 4.7 0 0 1-1.57-.26c-17.14-5.82-31.73-17-43.36-33.17-9.17-12.7-16.49-28.61-21.81-47.15a242.58 242.58 0 0 1-9.05-60.31 5.34 5.34 0 0 1 4.64-5.43c.4 0 40.32-4.38 68.13-25.79a4.92 4.92 0 0 1 6 0c27.8 21.41 67.72 25.75 68.12 25.79a5.35 5.35 0 0 1 4.64 5.43 242.58 242.58 0 0 1-9.05 60.31c-5.3 18.54-12.64 34.41-21.8 47.15-11.64 16.19-26.23 27.35-43.37 33.17a4.6 4.6 0 0 1-1.52.26ZM493.06 42a238.41 238.41 0 0 0 8.53 52.61c11 38.26 30.12 62.58 56.83 72.32 26.77-9.76 45.91-34.17 56.9-72.57A238.7 238.7 0 0 0 623.78 42c-11.51-1.78-41.34-7.81-65.36-24.94-24.03 17.15-53.85 23.17-65.36 24.94Z"
        transform="translate(-5.43 -4.98)"
        style={{
          fill: "#36c6f4",
        }}
      />
      <path
        className="bc8c4e7a-0a4c-47d4-8d3b-876939d268e6"
        d="M584.71 44.57h19.43v19.77h-19.43Z"
        transform="translate(-5.43 -4.98)"
      />
      <path
        className="a9fe37ec-4cf7-43eb-a57d-cd438e81ae89"
        d="m548.82 60.86 11.42-16L576 56.49l-11.42 16ZM575.7 77.74l18.38-6.44 6.33 18.7L582 96.43Z"
        transform="translate(-5.43 -4.98)"
      />
      <path
        className="bc8c4e7a-0a4c-47d4-8d3b-876939d268e6"
        d="M520.51 64.17H540v19.76h-19.49Z"
        transform="translate(-5.43 -4.98)"
      />
      <path
        className="a9fe37ec-4cf7-43eb-a57d-cd438e81ae89"
        d="M520.51 87.09H540v19.77h-19.49Z"
        transform="translate(-5.43 -4.98)"
      />
      <path
        className="bc8c4e7a-0a4c-47d4-8d3b-876939d268e6"
        d="M542.89 87.09h19.44v19.77h-19.44ZM520.51 109.84H540v19.77h-19.49Z"
        transform="translate(-5.43 -4.98)"
      />
      <path
        className="a9fe37ec-4cf7-43eb-a57d-cd438e81ae89"
        d="M542.89 109.84h19.44v19.77h-19.44Z"
        transform="translate(-5.43 -4.98)"
      />
      <path
        className="bc8c4e7a-0a4c-47d4-8d3b-876939d268e6"
        d="M565.43 109.84h19.44v19.77h-19.44Z"
        transform="translate(-5.43 -4.98)"
        style={{
          fill: theme.palette.primary.main,
        }}
      />
      <path
        d="M57.46 146.1h-11.4L23.32 80.63l1.49-2.19h19.37v-45H16.29V146.1H5.43V23.84h40.94A8.29 8.29 0 0 1 55 32.51v46.4q0 8.84-11.87 8.83c-.78 0-1.93 0-3.44-.12s-2.61-.11-3.28-.11Q47 116.73 57.46 146.1ZM113.87 146.1H73.71V23.84h39.69v9.6H84.57v45.63h26.25v9.77H84.57v47.65h29.3ZM180.28 137.43a8.31 8.31 0 0 1-8.68 8.67h-32.11a8.29 8.29 0 0 1-8.67-8.67V32.51a8.29 8.29 0 0 1 8.67-8.67h32.11a8.31 8.31 0 0 1 8.68 8.67v28.43h-10.94v-27.5h-27.66v103.05h27.66V94.23h-13.05v-9.38h24ZM211.14 146.1h-10.86V23.84h10.86ZM279.18 137.43a8.52 8.52 0 0 1-2.38 6.25 8.35 8.35 0 0 1-6.21 2.42h-30.78a8.6 8.6 0 0 1-6.29-2.42 8.38 8.38 0 0 1-2.46-6.25v-31h10.86v30.08h26.4v-29.16l-33.75-40.78a14.2 14.2 0 0 1-3.51-9.3V32.51a8.38 8.38 0 0 1 2.46-6.25 8.56 8.56 0 0 1 6.29-2.42h30.78a8.31 8.31 0 0 1 6.21 2.42 8.52 8.52 0 0 1 2.38 6.25v28.43h-10.86v-27.5h-26.4v25.79l33.9 40.77a14 14 0 0 1 3.36 9.14ZM340.51 33.44H321.6V146.1h-10.86V33.44h-19.06v-9.6h48.83ZM405.28 146.1h-11.41l-22.73-65.47 1.48-2.19H392v-45h-27.9V146.1h-10.86V23.84h40.94a8.29 8.29 0 0 1 8.67 8.67v46.4q0 8.84-11.87 8.83c-.78 0-1.93 0-3.44-.12s-2.6-.11-3.28-.11q10.55 29.22 21.02 58.59ZM470.67 24.54 455.82 74q-2.82 7.41-7.42 24.76v47.34h-10.86V98.76a86.36 86.36 0 0 0-3.36-12.5Q430.43 75 430.12 74l-14.84-49.46v-.7h11L443.09 85l16.64-61.17h10.94ZM692.17 146.1h-10.78l-4.29-27.19h-24.3l-4.3 27.19H638v-.31l21.8-122.27h10.7Zm-16.48-36.8L665 43.05l-10.79 66.25ZM757.25 124q0 9.69-6.36 15.9a22.08 22.08 0 0 1-16.06 6.21h-28V23.84h28q9.76 0 16.09 6.21T757.25 46Zm-10.86-1.09v-76a13 13 0 0 0-3.79-9.72 13.37 13.37 0 0 0-9.8-3.71h-15.08v103h14.45a14.59 14.59 0 0 0 10.28-3.63q3.95-3.62 3.94-9.95ZM836.78 146.1h-10.23V68.68c0-1.09.36-4.3 1.09-9.61l-20 75.55h-2.18l-20-75.55c.72 5.37 1.09 8.57 1.09 9.61v77.42h-10.23V23.84h10.07l19.85 79.53a30.25 30.25 0 0 1 .31 3.51 28.63 28.63 0 0 1 .31-3.51l19.85-79.53h10.07ZM867.8 146.1h-10.86V23.84h10.86ZM938.58 146.1h-7.5l-32.66-89.92v89.92h-10.23V23.84h8.13l32 88.36V23.84h10.23Z"
        transform="translate(-5.43 -4.98)"
        style={{
          fill: logoTheme.palette.primary.main,
        }}
      />
    </g>
  </svg>


}

export default Logo;