class PageUpdater {
    constructor(webSocketUrl) {
        this.socket = new WebSocket(webSocketUrl);
        this.socket.onmessage = this.handleMessage.bind(this);
        this.socket.onopen = (event) => {
            console.log('WebSocket is open now.');
        };
    }

    handleMessage(event) {
        const data = JSON.parse(event.data);
        if (Array.isArray(data.updates)) {
            data.updates.forEach(update => this.applyUpdate(update));
        } else {
            this.applyUpdate(data);
        }
    }

    applyUpdate(update) {
        if (update.id && update.html) {
            this.updateElementById(update.id, update.html);
        } else if (update.class && update.innerHTML) {
            this.updateInnerHTMLByClass(update.class, update.innerHTML);
        } else {
            console.log("invalid ws update:", update);
        }
    }

    applyUpdate(update) {
        switch (update.command) {
            case 'setElement':
                if (update.id) {
                    this.updateElementById(update.id, update.html);
                    return;
                }break;
            case 'setInnerHTML':
                if (update.id) {
                    this.setInnerHTMLById(update.id, update.html);
                    return;
                } else if (update.class) {
                    this.setInnerHTMLByClass(update.class, update.html);
                    return;
                }break;
            case 'setClassNames':
                if (update.id) {
                    this.setClassNamesById(update.id, update.classNames);
                    return;
                }break;
            case 'toggleClassNames':
                if (update.id && update.classNames) {
                    this.toggleClassNamesById(update.id, update.classNames);
                    return;
                } else if (update.class && update.classNames) {
                    this.toggleClassNamesByClass(update.class, update.classNames);
                    return;
                }break;
            default:
                console.error('Unknown command:', update);
                return;
        }
        console.error('Invalid command data:', update);
    }

    setClassNamesById(id, classNames) {
        const element = document.getElementById(id);
        if (element) {
            element.className = classNames;
        }
    }

    toggleClassNamesById(id, classNames) {
        const element = document.getElementById(id);
        if (element) {
            Array.from(classNames).forEach(className => {
                element.classList.toggle(className);
            });
        }
    }

    toggleClassNamesByClass(className, classNames) {
        const elements = document.getElementsByClassName(className);
        Array.from(elements).forEach(element => {
            Array.from(classNames).forEach(className => {
                element.classList.toggle(className);
            });
        });
    }

    setInnerHTMLById(id, html) {
        const element = document.getElementById(id);
        if (element) {
            element.innerHTML = html;
        }
    }

    setInnerHTMLByClass(className, html) {
        const elements = document.getElementsByClassName(className);
        Array.from(elements).forEach(element => {
            element.innerHTML = html;
        });
    }

    updateElementById(id, html) {
        const target = document.getElementById(id);
        if (target) {
            // creates a temporary container to parse the HTML
            const tempContainer = document.createElement('div');
            tempContainer.innerHTML = html;

            // Extract the new element with the matching ID
            let newElement = tempContainer.querySelector(`#${id}`);
            if (!newElement) {
                console.log("newly rendered element does not match expected id:", id, newElement);
                return;
            }
            // apply changes
            target.className = newElement.className
            target.style.cssText = newElement.style.cssText;
            Array.from(newElement.attributes).forEach(attr => {
                target.setAttribute(attr.name, attr.value);
            });
            target.innerHTML = newElement.innerHTML;
        }
    }
}


