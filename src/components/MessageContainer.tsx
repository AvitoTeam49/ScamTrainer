const MessageContainer = () => {
    return (
        <div className="messages-container">
            <div className="message-wrapper other">
                <span className="message-time">15:36</span>
                <div className="message-bubble">
                    Привет, хотел бы купить комплект!
                </div>
            </div>

            <div className="message-wrapper own">
                <span className="message-time">15:37</span>
                <div className="message-bubble">
                    Здравствуйте, хорошо.
                </div>
            </div>
        </div>
    );
};

export default MessageContainer;