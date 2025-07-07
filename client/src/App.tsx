import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import SearchPage from './pages/SearchPage';

function App() {
  return (
    <Router>
      <Routes>
          <Route path="/" element={<SearchPage />} />
        </Routes>
    </Router>
  );
}

export default App;