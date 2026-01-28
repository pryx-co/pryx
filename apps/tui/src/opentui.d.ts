import "@opentui/core";
import { JSX } from "solid-js";

declare module "@opentui/core" {
    interface BoxOptions {
        children?: JSX.Element;
    }
    interface TextOptions {
        children?: JSX.Element | string | number | (string | number)[];
    }
    interface InputOptions {
        onChange?: (value: string) => void;
        onSubmit?: (value: string) => void;
    }
}

declare global {
    namespace JSX {
        interface IntrinsicElements {
            box: any;
            text: any;
            input: any;
        }
    }
}
