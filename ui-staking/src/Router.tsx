import { createBrowserRouter, createRoutesFromElements, Route, RouterProvider } from "react-router-dom";
import { Landing } from "./pages/landing/Landing.tsx";
import { App } from "./pages/pool/Pool.tsx";
import React from "react";
import { Stake } from "./pages/stake/Stake.tsx";

const Root = (
  <>
    <Route path="/" element={<Landing />} />
    <Route path="/pool/:poolId" element={<App />} />
    <Route path="/pool/:poolId/stake" element={<Stake />} />
  </>
);

const router = createBrowserRouter(createRoutesFromElements(Root));

export const Router = () => {
  return (
    <React.Suspense fallback={<>Loading...</>}>
      <RouterProvider router={router} />
    </React.Suspense>
  );
};
