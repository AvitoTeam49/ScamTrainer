import { type FC, useContext } from "react";
import { Context } from "../main.tsx";
import { observer } from "mobx-react-lite";

const Profile: FC = observer(() => {
    const {menuOpen, user} = useContext(Context);

    return (
        <div className="profile-section">

            <button
                className="menu-toggle fixed"
                onClick={menuOpen.setTrue}
            >
                ☰
            </button>

            <h1>
                Профиль
            </h1>

            <div className="profile-card">

                <p>
                    <strong>
                        Username:
                    </strong>{" "}
                    {user.user?.username ?? "—"}
                </p>

                <p>
                    Ваш счет:{" "}
                    <span className="stat-number">
                        {user.user?.score ?? 0}
                    </span>
                </p>

            </div>

        </div>
    );
});

export default Profile;