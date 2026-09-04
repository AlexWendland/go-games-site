import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { useAuth, type AuthState } from "../context/AuthContext";
import { Logo, Person } from "./icons";

function AccountButton({ auth }: { auth: AuthState }) {
  const linkClass =
    "flex items-center gap-3 transition-all duration-200 hover:scale-105";

  if (auth.status === "authenticated") {
    return (
      <Link to="/account" className={`${linkClass} text-orange-200 hover:text-orange-400`}>
        <div className="text-right hidden sm:block">
          <div className="text-base font-semibold leading-tight">
            {auth.user.display_name}
          </div>
          <div className="text-xs text-gray-400 leading-tight">
            {auth.user.user_id}
          </div>
        </div>
        <Person size={32} />
      </Link>
    );
  }

  return (
    <Link to="/account" className={`${linkClass} text-gray-400 hover:text-orange-400`}>
      <span className="hidden sm:block text-sm">Sign in</span>
      <Person size={32} />
    </Link>
  );
}

export function NavBar() {
  const { auth } = useAuth();

  // Hide navbar when scrolling down, reveal when scrolling up.
  const [visible, setVisible] = useState(true);
  const lastScrollY = useRef(0);

  useEffect(() => {
    const handleScroll = () => {
      const currentScrollY = window.scrollY;
      const scrollingDown =
        currentScrollY > lastScrollY.current && currentScrollY > 50;
      const scrollingUp = currentScrollY < lastScrollY.current;

      if (scrollingDown) setVisible(false);
      if (scrollingUp) setVisible(true);

      lastScrollY.current = currentScrollY;
    };

    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  return (
    <nav
      className={`fixed top-0 left-0 right-0 z-50 transform transition-transform duration-300 ${
        visible ? "translate-y-0" : "-translate-y-full"
      } bg-gray-800 shadow-lg shadow-black/40 py-4`}
    >
      <div className="flex items-center justify-between">
        {/* Logo — pinned to the far left, outside the page column */}
        <Link to="/" className="flex items-center gap-3 text-orange-500 pl-4 transition-all duration-200 hover:scale-105 hover:text-orange-400">
          <Logo size={40} />
          <span className="hidden sm:block font-bold text-3xl">
            Alex's Games
          </span>
        </Link>

        <div className="pr-4">
          <AccountButton auth={auth} />
        </div>
      </div>
    </nav>
  );
}
