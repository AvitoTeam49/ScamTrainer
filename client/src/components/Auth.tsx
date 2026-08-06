import {type FC, useState} from "react";
import {useNavigate} from "react-router-dom";

const Auth:FC = () => {

    const [isPasswordLoginVisible, setIsPasswordLoginVisible] = useState<boolean>(false);
    const [isPasswordRegisterVisible, setIsPasswordRegisterVisible] = useState<boolean>(false);
    const [loginEmail, setLoginEmail] = useState<string>("");
    const [loginPassword, setLoginPassword] = useState<string>("");
    const [registerEmail, setRegisterEmail] = useState<string>("");
    const [registerPassword, setRegisterPassword] = useState<string>("");
    const navigate = useNavigate();

    return (
        <div className="main auth">
            <div className="header-logo">
                <div className="logo-text" onClick={() => navigate("/auth")}>
                    <span className="logo-icon">
                        <span className="logo-dot dot-blue"></span>
                        <span className="logo-dot dot-red"></span>
                        <span className="logo-dot dot-green"></span>
                    </span>
                    Avito
                </div>
            </div>
            <div className="main-content-auth">
                <div className="auth-wrapper">

                    <div className="auth-column">
                        <h2 className="auth-title">Вход</h2>

                        <div className="input-group">
                            <div className="input-wrapper-with-icon">
                                <input
                                    type="text"
                                    placeholder="Почта"
                                    value={loginEmail}
                                    onChange={(e) => setLoginEmail(e.target.value)}
                                    className="input-field"
                                />

                                {loginEmail.length > 0 && (
                                    <button
                                        className="clear-btn"
                                        type="button"
                                        onClick={() => setLoginEmail("")}
                                    >
                                        ×
                                    </button>
                                )}
                            </div>
                        </div>

                        <div className="input-group">
                            <div className="password-wrapper">
                                <input
                                    type={isPasswordLoginVisible ? "text" : "password"}
                                    placeholder="Пароль"
                                    className="input-field"
                                    value={loginPassword}
                                    onChange={(e) => setLoginPassword(e.target.value)}
                                />

                                {loginPassword.length > 0 && (
                                    <button
                                        className="toggle-password"
                                        type="button"
                                        onClick={() => setIsPasswordLoginVisible(!isPasswordLoginVisible)}
                                    >
                                        {isPasswordLoginVisible ? (
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                                                <circle cx="12" cy="12" r="3"></circle>
                                            </svg>
                                        ) : (
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
                                                <line x1="1" y1="1" x2="23" y2="23"></line>
                                            </svg>
                                        )}
                                    </button>
                                )}
                            </div>
                        </div>

                        <button className="action-btn">Войти</button>
                    </div>

                    <div className="auth-column">
                        <h2 className="auth-title">Регистрация</h2>

                        <div className="input-group">
                            <div className="input-wrapper-with-icon">
                                <input
                                    type="text"
                                    placeholder="Почта"
                                    value={registerEmail}
                                    onChange={(e) => setRegisterEmail(e.target.value)}
                                    className="input-field"
                                />
                                {registerEmail.length > 0 && (
                                    <button
                                        className="clear-btn"
                                        type="button"
                                        onClick={() => setRegisterEmail("")}
                                    >
                                        ×
                                    </button>
                                )}
                            </div>
                        </div>

                        <div className="input-group">
                            <div className="password-wrapper">
                                <input
                                    type={isPasswordRegisterVisible ? "text" : "password"}
                                    placeholder="Пароль"
                                    className="input-field"
                                    value={registerPassword}
                                    onChange={(e) => setRegisterPassword(e.target.value)}
                                />

                                {registerPassword.length > 0 && (
                                    <button
                                        className="toggle-password"
                                        type="button"
                                        onClick={() => setIsPasswordRegisterVisible(!isPasswordRegisterVisible)}
                                    >
                                        {isPasswordRegisterVisible ? (
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                                                <circle cx="12" cy="12" r="3"></circle>
                                            </svg>
                                        ) : (
                                            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                                                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
                                                <line x1="1" y1="1" x2="23" y2="23"></line>
                                            </svg>
                                        )}
                                    </button>
                                )}
                            </div>
                        </div>

                        <button className="action-btn">Зарегистрироваться</button>
                    </div>

                </div>
            </div>
        </div>


    );
};

export default Auth;