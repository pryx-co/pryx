import { Box, Text } from "@opentui/core";

export default function AppHeader() {
    return (
        <Box flexDirection="column" alignItems="center" marginBottom={1} borderStyle="single" borderColor="cyan" padding={1}>
            <Text color="cyan" bold>
                {`
  ____                  
 |  _ \\ _ __ _   _ __  __
 | |_) | '__| | | |\\ \\/ /
 |  __/| |  | |_| | >  < 
 |_|   |_|   \\__, |/_/\\_\\
             |___/       
`}
            </Text>
            <Text color="gray">  AI-Powered Agentic Coding Environment  </Text>
        </Box>
    );
}
