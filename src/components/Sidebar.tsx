import {observer} from "mobx-react-lite";
import isMenuOpen from "../store/isMenuOpen.ts";

const Sidebar = observer(() => {

    return (
        <div className={`sidebar ${isMenuOpen.menu ? 'open' : ''}`}>
            <div className="logo-container">
                <div className="logo-text">
                <span className="logo-icon">
                    <span className="logo-dot dot-blue"></span>
                    <span className="logo-dot dot-red"></span>
                    <span className="logo-dot dot-green"></span>
                </span>
                    Avito
                </div>
                <button className="close-sidebar-btn" onClick={isMenuOpen.setFalse}>✕</button>
            </div>

            <button className="new-chat-btn">
                <span>+</span> Новый чат
            </button>

            <ul className="user-list">
                <li className="user-item">User 1</li>
                <li className="user-item">User 2</li>
                <li className="user-item">User 3</li>
                <li className="user-item">User 4</li>
                <li className="user-item">User 5</li>
            </ul>
        </div>
    );
});

export default Sidebar;