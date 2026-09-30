import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-3 bg-base px-4 text-center">
      <p className="font-mono text-[10px] uppercase tracking-[0.14em] text-faint">
        Error 404
      </p>
      <h1 className="text-2xl text-ink">This page is off the map.</h1>
      <p className="max-w-sm text-sm text-muted">
        The view you were looking for does not exist or was moved.
      </p>
      <Link to="/" className="mt-2 text-sm text-accent hover:underline">
        Back to the landing page
      </Link>
    </div>
  );
}
