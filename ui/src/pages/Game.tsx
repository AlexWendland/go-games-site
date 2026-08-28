import { useParams } from "react-router-dom";

export default function Game() {
  const { id } = useParams();
  return <h1 className="text-2xl text-red-800">Game {id}</h1>;
}
