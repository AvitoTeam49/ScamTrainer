import {type FC, useContext, useState} from "react";
import {observer} from "mobx-react-lite";
import {useNavigate} from "react-router-dom";
import {Context} from "../main.tsx";

const NewChatArea:FC = observer(() => {

    const {chat, menuOpen, user, scenario} = useContext(Context)
    const [selectedRole, setSelectedRole] = useState<string | null>(null)
    const [selectedDifficulty, setSelectedDifficulty] = useState<number | null>(null)
    const [isLoading, setIsLoading] = useState<boolean>(false)
    const [createError, setCreateError] = useState<string>("")
    const navigate = useNavigate()

    const handleCreateChat = async () => {
        if (isLoading) {return;}

        setIsLoading(true);
        setCreateError("");

        try {
            const difficulty = selectedDifficulty ?? 0;

            const role = selectedRole ?? "buyer";

            const scenarioResult = await scenario.getScenarios(difficulty);

            if (!scenarioResult.success || !scenarioResult.scenarios) {
                setCreateError("Не удалось загрузить сценарии");
                return;
            }

            const selectedScenario = scenarioResult.scenarios.find(item => item.role === role);

            if (!selectedScenario) {
                setCreateError("Для этой сложности и роли пока нет сценариев");
                return;
            }

            const chatResult = await chat.createChat(selectedScenario.id, selectedScenario.title);

            if (!chatResult.success || !chatResult.chat) {
                setCreateError("Не удалось создать чат");
                return;
            }

            navigate(`/chat/${chatResult.chat.id}`);

        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="main-content">

            <button
                className="menu-toggle fixed"
                onClick={menuOpen.setTrue}
            >
                ☰
            </button>

            <button className="header-username" onClick={() => navigate("/profile")}>{user.user?.username}</button>


            <h1 className="title">Создать новый чат</h1>

            <div className="label">Выберите роль</div>
            <div className="button-group">
                <button className={`option-btn ${selectedRole === 'buyer' ? 'active' : ''}`}
                onClick={() => setSelectedRole('buyer')}>
                    Покупатель</button>
                <button className={`option-btn ${selectedRole === 'seller' ? 'active' : ''}`}
                onClick={() => setSelectedRole('seller')}>
                    Продавец</button>
            </div>

            <div className="label">Выберите сложность</div>
            <div className="button-group">
                <button className={`option-btn ${selectedDifficulty === 0 ? "active" : ""}`}
                onClick={() => setSelectedDifficulty(0)}>
                    Легкая</button>
                <button className={`option-btn ${selectedDifficulty === 1 ? "active" : ""}`}
                        onClick={() => setSelectedDifficulty(1)}>
                    Средняя</button>
                <button className={`option-btn ${selectedDifficulty === 2 ? "active" : ""}`}
                        onClick={() => setSelectedDifficulty(2)}>
                    Сложная</button>
            </div>

            <button className="create-btn" onClick={handleCreateChat} disabled={isLoading}>Создать</button>

            {createError && (
                <div className="create-error">{createError}</div>
            )}
        </div>
    );
});

export default NewChatArea;