import {useState} from "react";
import isMenuOpen from "../store/isMenuOpen.ts";


const NewChatArea = () => {

    const [selectedRole, setSelectedRole] = useState<string | null>(null)
    const [selectedDifficulty, setSelectedDifficulty] = useState<string | null>(null)

    return (
        <div className="main-content">

            <button
                className="menu-toggle fixed"
                onClick={isMenuOpen.setTrue}
            >
                ☰
            </button>


            <h1 className="title">Создать новый чат</h1>

            <div className="label">Выберите роль</div>
            <div className="button-group">
                <button
                    className={`option-btn ${selectedRole === 'buyer' ? 'active' : ''}`}
                onClick={() => setSelectedRole('buyer')}>
                    Покупатель</button>
                <button
                    className={`option-btn ${selectedRole === 'seller' ? 'active' : ''}`}
                onClick={() => setSelectedRole('seller')}>
                    Продавец</button>
            </div>

            <div className="label">Выберите сложность</div>
            <div className="button-group">
                <button className={`option-btn ${selectedDifficulty === "easy" ? "active" : ""}`}
                onClick={() => setSelectedDifficulty('easy')}>
                    Легкая</button>
                <button className={`option-btn ${selectedDifficulty === "medium" ? "active" : ""}`}
                        onClick={() => setSelectedDifficulty('medium')}>
                    Средняя</button>
                <button className={`option-btn ${selectedDifficulty === "hard" ? "active" : ""}`}
                        onClick={() => setSelectedDifficulty('hard')}>
                    Сложная</button>
            </div>

            <button className="create-btn">Создать</button>
        </div>
    );
};

export default NewChatArea;