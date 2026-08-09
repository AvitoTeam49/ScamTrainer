import {type FC, useContext} from "react";
import {Context} from "../main.tsx";
import {observer} from "mobx-react-lite";
import type {ITopUser} from "../types/types.tsx";

const LeaderBoard: FC = observer(() => {
    const {user} = useContext(Context);
    
    const topUsers = user.leaderboard?.top_users || [];

    if (topUsers.length === 0) {
        return (
            <div className="leaderboard-section">
                <h1>Лидеры</h1>
                <div className="loading-container">
                    <div className="spinner"></div>
                </div>
            </div>
        );
    }

    return (
        <div className="leaderboard-section">
            <h1>Лидеры</h1>

            <ul className="leaderboard-list">
                {topUsers.map((u: ITopUser) => {
                    return (
                        <li className="leader-item" key={u.rank}>
                            <div className="leader-info">
                                <span className="rank-badge">{u.rank}</span>
                                <span className="leader-username">{u.username}</span>
                            </div>
                            <span className="leader-score">{u.score}</span>
                        </li>
                    );
                })}
            </ul>
        </div>
    );
});

export default LeaderBoard;