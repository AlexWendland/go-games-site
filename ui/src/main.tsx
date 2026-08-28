import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import "./main.css";
import Home from "./pages/Home";
import Game from "./pages/Game";
import { BrowserRouter, Routes, Route } from "react-router-dom";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/game/:id" element={<Game />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
);
