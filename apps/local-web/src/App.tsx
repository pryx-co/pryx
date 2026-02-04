import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Dashboard from './components/Dashboard';
import SkillList from './components/skills/SkillList';
import Settings from './components/Settings';
import ChannelList from './components/channels/ChannelList';
import McpServerList from './components/mcp/McpServerList';
import PolicyList from './components/policy/PolicyList';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/skills" element={<SkillList />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/channels" element={<ChannelList />} />
        <Route path="/mcp" element={<McpServerList />} />
        <Route path="/policies" element={<PolicyList />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
