import {type FC, useContext, useState} from "react";
import {observer} from "mobx-react-lite";
import {useNavigate} from "react-router-dom";
import {Context} from "../main.tsx";
import type {IScenario} from "../types/types.tsx";


const NewChatArea:FC = observer(() => {

    const {chat, menuOpen, user, scenario} = useContext(Context)
    const [selectedRole, setSelectedRole] = useState<string | null>(null)
    const [selectedDifficulty, setSelectedDifficulty] = useState<number | null>(null)
    const [isLoading, setIsLoading] = useState<boolean>(false)
    const navigate = useNavigate()

    const handleCreateChat = async () => {
        setIsLoading(true)
        const scenario_result = await scenario.getScenarios(selectedDifficulty? selectedDifficulty : 1)
        if(!scenario_result.success){
            setIsLoading(false)
            return
        }
        const selectedScenario: IScenario = scenario.scenarios.find(
            item => item.role === (selectedRole ? selectedRole : "buyer")
        );

        if (!selectedScenario) {
            setIsLoading(false);
            return;
        }
        const scenario_id =  selectedScenario.id
        const scenario_title = selectedScenario.title

        const chat_result = await chat.createChat(scenario_id, scenario_title)
        if(!chat_result.success || !chat_result.chat){
            setIsLoading(false)
            return
        }

        navigate(`/chat/${chat_result.chat.id}`)
        setIsLoading(false)

    }

    return (
        <div className="main-content">

            <button
                className="menu-toggle fixed"
                onClick={menuOpen.setTrue}
            >
                ☰
            </button>

            <button className="header-username" onClick={() => navigate("/profile")}>{user.user.username}</button>


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
                <button className={`option-btn ${selectedDifficulty === 1 ? "active" : ""}`}
                onClick={() => setSelectedDifficulty(0)}>
                    Легкая</button>
                <button className={`option-btn ${selectedDifficulty === 2 ? "active" : ""}`}
                        onClick={() => setSelectedDifficulty(2)}>
                    Средняя</button>
                <button className={`option-btn ${selectedDifficulty === 3 ? "active" : ""}`}
                        onClick={() => setSelectedDifficulty(3)}>
                    Сложная</button>
            </div>

            <button className="create-btn" onClick={handleCreateChat} disabled={isLoading}>Создать</button>
        </div>
    );
});

export default NewChatArea;