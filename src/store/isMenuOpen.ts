import { makeAutoObservable } from 'mobx';

class isMenuOpen{
    menu = false
    constructor() {
        makeAutoObservable(this);
    }

    setTrue = ()=>{
        this.menu = true;
    }

    setFalse = () =>{
        this.menu = false;
    }
}

export default new isMenuOpen();