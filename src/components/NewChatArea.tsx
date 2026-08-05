import {use, useState} from "react";
import isMenuOpen from "../store/MenuOpen.ts";
import {observer} from "mobx-react-lite/src/observer.ts";
import Chat from "../store/Chat.ts";


const NewChatArea = observer(() => {

    const [selectedRole, setSelectedRole] = useState<string | null>(null)
    const [selectedDifficulty, setSelectedDifficulty] = useState<string | null>(null)
    const [counter, setCounter] = useState<number>(2)

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

            <button className="create-btn" onClick={() => {
                Chat.addNewChat({id: counter, role: selectedRole ? selectedRole : "buyer", difficulty: selectedDifficulty ? selectedDifficulty : "easy"})
                setCounter(prev => prev + 1)
            }}>Создать</button>
        </div>
    );
});

export default NewChatArea;