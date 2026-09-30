import type { ReactElement } from "react";
import {
  BrowserRouter,
  Navigate,
  Route,
  Routes,
  useLocation,
} from "react-router-dom";

import { useIsAuthenticated } from "@/features/auth/store";
import { DashboardLayout } from "@/layouts/DashboardLayout";
import { AnalyticsPage } from "@/pages/Analytics";
import { AudiencePage } from "@/pages/Audience";
import { DashboardPage } from "@/pages/Dashboard";
import { LandingPage } from "@/pages/Landing";
import { LoginPage } from "@/pages/Login";
import { NetworkPage } from "@/pages/Network";
import { NotFoundPage } from "@/pages/NotFound";
import { ProjectsPage } from "@/pages/Projects";
import { RegisterPage } from "@/pages/Register";
import { ReportsPage } from "@/pages/Reports";
import { SentimentPage } from "@/pages/Sentiment";
import { TimelinePage } from "@/pages/Timeline";
import { TrendsPage } from "@/pages/Trends";

function RequireAuth({ children }: { children: ReactElement }) {
  const authenticated = useIsAuthenticated();
  const location = useLocation();
  if (!authenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  return children;
}

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route
          path="/projects"
          element={
            <RequireAuth>
              <ProjectsPage />
            </RequireAuth>
          }
        />
        <Route
          path="/projects/:projectId"
          element={
            <RequireAuth>
              <DashboardLayout />
            </RequireAuth>
          }
        >
          <Route index element={<DashboardPage />} />
          <Route path="analytics" element={<AnalyticsPage />} />
          <Route path="timeline" element={<TimelinePage />} />
          <Route path="sentiment" element={<SentimentPage />} />
          <Route path="audience" element={<AudiencePage />} />
          <Route path="trends" element={<TrendsPage />} />
          <Route path="network" element={<NetworkPage />} />
          <Route path="reports" element={<ReportsPage />} />
        </Route>
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  );
}
