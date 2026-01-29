export default function AppHeader() {
    return (
        <box 
            flexDirection="column" 
            alignItems="center" 
            padding={1}
            backgroundColor="#0a0a0a"
        >
            <box flexDirection="row">
                <text fg="#00ffff">{'  ____                  '}</text>
            </box>
            <box flexDirection="row">
                <text fg="#00ffff">{" |  _ \\ _ __ _   _ __  __"}</text>
            </box>
            <box flexDirection="row">
                <text fg="#00ffff">{" | |_) | '__| | | |\\ \\/ /"}</text>
            </box>
            <box flexDirection="row">
                <text fg="#00ffff">{" |  __/| |  | |_| | \u003e  \u003c "}</text>
            </box>
            <box flexDirection="row">
                <text fg="#00ffff">{" |_|   |_|   \\__, |/_/\\_\\"}</text>
            </box>
            <box flexDirection="row">
                <text fg="#00ffff">{'              |___/       '}</text>
            </box>
            <box marginTop={1}>
                <text fg="#808080">Autonomous AI Agent for Any Task</text>
            </box>
        </box>
    );
}
