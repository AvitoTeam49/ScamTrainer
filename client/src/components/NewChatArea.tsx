import {type FC, useContext, useState} from "react";
import {observer} from "mobx-react-lite";
import {useNavigate} from "react-router-dom";
import {Context} from "../main.tsx";


const NewChatArea:FC = observer(() => {

    const {chat, menuOpen} = useContext(Context)

    const [selectedRole, setSelectedRole] = useState<string | null>(null)
    const [selectedDifficulty, setSelectedDifficulty] = useState<string | null>(null)
    let old_id = chat.chats.length > 0 ? chat.chats[chat.chats.length - 1].id : 1
    const navigate = useNavigate()

    return (
        <div className="main-content">

            <button
                className="menu-toggle fixed"
                onClick={menuOpen.setTrue}
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
                chat.addNewChat({id: old_id + 1, role: selectedRole ? selectedRole : "buyer", difficulty: selectedDifficulty ? selectedDifficulty : "easy"})
                navigate(`/chat/${old_id + 1}`)
            }}>Создать</button>
        </div>
    );
});

export default NewChatArea;