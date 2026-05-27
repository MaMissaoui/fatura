// Wails v2 runtime type declarations.
export function WindowSetTitle(title: string): void;
export function WindowFullscreen(): void;
export function WindowUnfullscreen(): void;
export function BrowserOpenURL(url: string): void;
export function EventsOn(event: string, callback: (...data: any[]) => void): () => void;
export function EventsOff(event: string, ...callbacks: ((...data: any[]) => void)[]): void;
export function EventsEmit(event: string, ...data: any[]): void;
