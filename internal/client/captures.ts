import { fetchOpt, fetchOptJson, fetchOptForm, date } from "./lib";
import navigator from "./navigator";
import { detached } from "./params";

interface EventOpt {
    preventDefault?: boolean;
    stopPropagation?: boolean;
}
interface InputOpt {
    excludeValue: boolean;
}

function applyEventOpt(event: Event, opt: EventOpt): void {
    if (opt.preventDefault) {
        event.preventDefault();
    }
    if (opt.stopPropagation) {
        event.stopPropagation();
    }
}

interface InputValues {
    name: string | null;
    value: string;
    number: number | null;
    date: string | null;
    selected: string[];
    checked: boolean;
}

function getInputValues(input: HTMLInputElement | HTMLSelectElement): InputValues {
    const value = input.value;
    let number: number | null = (input as HTMLInputElement).valueAsNumber;
    if (isNaN(number)) {
        number = null;
    }

    let dateValue: string | null = null;
    const valueAsDate = (input as HTMLInputElement).valueAsDate;
    if (valueAsDate) {
        dateValue = valueAsDate.toISOString();
    }

    let selected: string[] = [];
    if ('selectedOptions' in input && input.selectedOptions) {
        selected = Array.from(input.selectedOptions).map(option => option.value);
    }
    const checked = 'checked' in input ? input.checked === true : false;
    const name = input.name || null;

    return { name, value, number, date: dateValue, selected, checked };
}

export default {
    default(data: any) {
        return fetchOpt(data);
    },

    json(data: any) {
        return fetchOptJson(data);
    },

    link(event: MouseEvent, opt: EventOpt) {
        opt.preventDefault = true;
        applyEventOpt(event, opt);
        const href = (event.currentTarget as HTMLAnchorElement).href;
        if (href && !detached) {
            navigator.push(href);
        }
        return {};
    },

    focus(event: FocusEvent) {
        const obj = {
            type: event.type,
            timestamp: date(new Date()),
        };
        return fetchOptJson(obj);
    },
    focus_io(event: FocusEvent, opt: EventOpt) {
        applyEventOpt(event, opt);
        const obj = {
            type: event.type,
            timestamp: date(new Date()),
        };
        return fetchOptJson(obj);
    },

    keyboard(event: KeyboardEvent, opt: EventOpt) {
        applyEventOpt(event, opt);
        const obj = {
            type: event.type,
            key: event.key,
            code: event.code,
            repeat: event.repeat,
            altKey: event.altKey,
            ctrlKey: event.ctrlKey,
            shiftKey: event.shiftKey,
            metaKey: event.metaKey,
            timestamp: date(new Date()),
        };
        return fetchOptJson(obj);
    },

    pointer(event: PointerEvent, opt: EventOpt) {
        applyEventOpt(event, opt);
        const obj = {
            type: event.type,
            pointerId: event.pointerId,
            width: event.width,
            height: event.height,
            pressure: event.pressure,
            tangentialPressure: event.tangentialPressure,
            tiltX: event.tiltX,
            tiltY: event.tiltY,
            twist: event.twist,
            buttons: event.buttons,
            button: event.button,
            pointerType: event.pointerType,
            isPrimary: event.isPrimary,
            clientX: event.clientX,
            clientY: event.clientY,
            screenX: event.screenX,
            screenY: event.screenY,
            pageX: event.pageX,
            pageY: event.pageY,
            timestamp: date(new Date()),
        };
        return fetchOptJson(obj);
    },
    input(event: InputEvent, opt: InputOpt) {
        return fetchOptJson({
            type: event.type,
            data: event.data,
            ...opt.excludeValue === true ? {} : getInputValues(event.target as HTMLInputElement | HTMLSelectElement),
            timestamp: date(new Date()),
        });
    },
    change(event: Event) {
        return fetchOptJson({
            type: event.type,
            ...getInputValues(event.target as HTMLInputElement | HTMLSelectElement),
            timestamp: date(new Date()),
        });
    },

    submit(event: SubmitEvent) {
        applyEventOpt(event, { preventDefault: true });
        const form = event.target as HTMLFormElement;
        const formData = new FormData(form);
        return fetchOptForm(formData);
    }
};

