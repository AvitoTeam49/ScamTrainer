import { makeAutoObservable } from 'mobx';

class MenuOpen {
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

export default MenuOpen;