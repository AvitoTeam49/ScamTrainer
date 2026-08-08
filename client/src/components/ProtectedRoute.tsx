import type {ReactNode} from "react";
import {Navigate, Outlet} from "react-router-dom";

interface ProtectedRouteProps {
    isAllowed: boolean;
    children?: ReactNode;
}

const ProtectedRoute = ({isAllowed, children}: ProtectedRouteProps) => {

    if(!isAllowed) {
        return <Navigate to="/auth" replace/>;
    }

    return children ? <>{children}</> : <Outlet />;
};

export default ProtectedRoute;