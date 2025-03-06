// import './assets/main.css'
import ReactDOM from 'react-dom/client';
import App from './App';
import ReactModal from 'react-modal';

const root = document.getElementById('root');
if (!root) {
  throw new Error('Root element not found');
}
ReactDOM.createRoot(root).render(
  // <React.StrictMode>
  <App />,
  // </React.StrictMode>
);

ReactModal.setAppElement(root);
