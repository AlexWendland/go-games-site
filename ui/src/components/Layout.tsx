import { Outlet } from "react-router-dom";
import { NavBar } from "./NavBar";

export default function Layout() {
  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <NavBar />
      {/* pt-20 clears the fixed navbar */}
      <div className="relative flex flex-col pt-20">
        <main className="container mx-auto max-w-5xl px-4 pt-8 flex-grow">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
