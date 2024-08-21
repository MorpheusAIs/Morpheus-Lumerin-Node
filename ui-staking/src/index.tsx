import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./App.tsx";
import { App2 } from "./App2.tsx";

const rootElem = document.getElementById("root");
if (!rootElem) {
	throw new Error("Root element not found");
}

// ReactDOM.createRoot(rootElem).render(<App />);
ReactDOM.createRoot(rootElem).render(<App2 />);
