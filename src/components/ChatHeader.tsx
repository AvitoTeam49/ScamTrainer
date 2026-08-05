interface ChatHeaderProps {
    toggleMenu: () => void
}

const ChatHeader = ({toggleMenu}: ChatHeaderProps) => {
    return (
        <div className="chat-header">
            <button className="menu-toggle" onClick={toggleMenu}>☰</button>
            <div className="user-avatar">U</div>
            <div className="chat-info">
                <div className="chat-title">Комплект GPU</div>
                <div className="chat-price">500₽</div>
            </div>
        </div>
    );
};

export default ChatHeader;