import {type FC, useContext} from "react";
import {Context} from "../main.tsx";
import {observer} from "mobx-react-lite";

const Profile:FC = observer(() => {
    const {menuOpen, user} = useContext(Context)

    return (
        <div className="profile-section">
            <button
                className="menu-toggle fixed"
                onClick={menuOpen.setTrue}
            >
                ☰
            </button>

            <h1>Профиль</h1>

            <div className="profile-card">
                <p><strong>Username:</strong>{user.user?.username}</p>
                <p>Успешно пройдено сценариев: <span className="stat-number">{user.progress?.scams_detected}</span></p>
                <p>Провалено сценариев: <span className="stat-number">{user.progress?.failed_attempts}</span></p>
                <p>Ваш счет: <span className="stat-number">{user.user?.score}</span></p>
                <p>Ваш ранг: <span className="stat-number">{user.leaderboard?.user_rank}</span></p>
            </div>
        </div>
    );
});

export default Profile;