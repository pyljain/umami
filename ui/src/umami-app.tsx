import { useEffect, useState, useCallback } from 'react';
import { Plus, Play, Download, ArrowLeft, Pickaxe, CheckCircle, Eye, Edit2, Check, X, Send } from 'lucide-react';
import CreateAppModal from './CreateAppModal';

interface App {
  id: number;
  name: string;
  description: string;
  created: string;
  status: string;
  color: string;
}

interface Task {
  id: number | null;
  title: string;
  description: string;
  created: string;
  status: string;
  color: string;
  progress: number | null;
  completedDate: string;
}

interface Log {
  taskID: string;
  messages: {
    time: string;
    title: string;
    text: string;
  }[]
}

const colors = ["from-purple-500 to-pink-500", "from-blue-500 to-cyan-500", "from-green-500 to-emerald-500"]

const UmamiApp = () => {
  const [currentView, setCurrentView] = useState('home');
  const [selectedApp, setSelectedApp] = useState<App | null>(null);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const [apps, setApps] = useState<App[]>([]);
  const [tasks, setTasks] = useState<Record<string, Task[]>>({});
  const [logs, setLogs] = useState<Log>();
  const [showCreateAppModal, setShowCreateAppModal] = useState(false);

  useEffect(() => {
    fetchApps();
  }, []);

  async function fetchApps() {
    try {
      const response = await fetch('/api/v1/apps');
      const data = await response.json();
      setApps(data);
    } catch (error) {
      console.error('Error fetching apps:', error);
    }
  }

  async function fetchTasks(appId: number) {
    try {
      const response = await fetch(`/api/v1/apps/${appId}/tasks`);
      const data: Task[] = await response.json();
      console.log(data);
      // Bucket by status
      const tasks: Record<string, Task[]> = {
        authoring: [],
        inProgress: [],
        completed: []
      };
      data.forEach((task: Task) => {
        const status = task.status == "in-progress" ? "inProgress" : task.status;
        if (!tasks[status]) {
          tasks[status] = [];
        }
        tasks[status].push(task);
      });

      setTasks(tasks);
    } catch (error) {
      console.error('Error fetching tasks:', error);
    }
  }

  async function fetchLogs(taskId: number) {
    try {
      const response = await fetch(`/api/v1/apps/${selectedApp?.id}/tasks/${taskId}/logs`);
      const data: Log = await response.json();
      setLogs(data);

      // Start the log stream
      const ws = new WebSocket(`ws://${location.host}/api/v1/apps/${selectedApp?.id}/tasks/${taskId}/logs/ws`);
      ws.onerror = function(e) {
        console.error(`Error in websocket connection ${e}`)
      }
      ws.onopen = function() {
        console.log("Socket opened")
      }
      ws.onclose = function() {
        console.log('Socket closed')
      }
      ws.onmessage = function(message) {
        const data: Log = JSON.parse(message.data);
        setLogs(data);
      }
    } catch (error) {
      console.error('Error fetching logs:', error);
    }
  }

  const createApp = useCallback(async (name: string, description: string) => {
    try {
      const response = await fetch('/api/v1/apps', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name,
          description,
        }),
      });

      if (response.ok) {
        await fetchApps(); // Refresh the apps list
      } else {
        const errorText = await response.text();
        alert(`Failed to create app: ${errorText}`);
        throw new Error(errorText);
      }
    } catch (error) {
      console.error('Error creating app:', error);
      alert('Failed to create app. Please try again.');
      throw error;
    }
  }, []);

  const handleCloseModal = useCallback(() => {
    setShowCreateAppModal(false);
  }, []);

  const downloadCode = async (appId: number) => {
    try {
      const response = await fetch(`/api/v1/apps/${appId}/download`);

      if (!response.ok) {
        throw new Error(`Failed to download code: ${response.statusText}`);
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `app-${appId}-code.zip`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('Error downloading code:', error);
      alert('Failed to download code. Please try again.');
    }
  };

  const AppsHomeScreen = () => (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      <div className="container mx-auto px-6 py-8">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-4xl font-bold text-slate-800 mb-2">Umami</h1>
            <p className="text-slate-600">AI-powered application builder</p>
          </div>
          <button 
            onClick={() => setShowCreateAppModal(true)}
            className="bg-gradient-to-r from-indigo-600 to-purple-600 text-white px-6 py-3 rounded-xl hover:from-indigo-700 hover:to-purple-700 transition-all duration-200 flex items-center gap-2 shadow-lg hover:shadow-xl"
          >
            <Plus size={20} />
            Create New App
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {apps.map((app, i) => (
            <div
              key={app.id}
              onClick={async () => {
                await fetchTasks(app.id);
                setSelectedApp(app);
                setCurrentView('app');
              }}
              className="bg-white rounded-2xl p-6 shadow-lg hover:shadow-xl transition-all duration-300 cursor-pointer group border border-slate-200 hover:border-slate-300"
            >
              <div className={`w-full h-32 bg-gradient-to-br ${colors[i%colors.length]} rounded-xl mb-4 relative overflow-hidden`}>
                <div className="absolute inset-0 bg-white/10 backdrop-blur-sm flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                  <Play className="text-white" size={32} />
                </div>
              </div>
              
              <h3 className="text-xl font-semibold text-slate-800 mb-2">{app.name}</h3>
              <p className="text-slate-600 text-sm mb-4 line-clamp-2">{app.description}</p>
              
              <div className="flex items-center justify-between">
                <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                  app.status === 'Active' ? 'bg-green-100 text-green-700' : 'bg-blue-100 text-blue-700'
                }`}>
                  {app.status}
                </span>
                <span className="text-xs text-slate-500">{app.created}</span>
              </div>
            </div>
          ))}
        </div>
        
        <CreateAppModal 
          isOpen={showCreateAppModal}
          onClose={handleCloseModal}
          onSubmit={createApp}
        />
      </div>
    </div>
  );

  const TaskCard = ({ task, column }: { task: Task, column: string }) => {
    const [isEditing, setIsEditing] = useState(false);
    const [editTitle, setEditTitle] = useState(task.title || '');
    const [editDescription, setEditDescription] = useState(task.description || '');

    const handleEdit = (e: React.MouseEvent) => {
      e.stopPropagation();
      setIsEditing(true);
    };

    const handleSave = async (e: React.MouseEvent) => {
      e.stopPropagation();
      try {
        let method = 'POST';
        let url = `/api/v1/apps/${selectedApp?.id}/tasks`;
        if (task.id) {
          method = 'PATCH';
          url = `/api/v1/apps/${selectedApp?.id}/tasks/${task.id}`;  
        }

        const response = await fetch(url, {
          method: method,
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            title: editTitle,
            description: editDescription,
          }),
        });
        
        if (response.ok) {
          // Update local state
          // setTasks(prev => ({
          //   ...prev,
          //   [column]: prev[column].map(t => 
          //     t.id === task.id 
          //       ? { ...t, name: editTitle, title: editTitle, description: editDescription }
          //       : t
          //   )
          // }));
          setIsEditing(false);
          fetchTasks(selectedApp?.id || 0);
        }
      } catch (error) {
        console.error('Error updating task:', error);
      }
    };


    const handleCancel = (e: React.MouseEvent) => {
      e.stopPropagation();
      setEditTitle(task.title ||'');
      setEditDescription(task.description || '');
      setIsEditing(false);
    };

    const handleSubmit = async (e: React.MouseEvent) => {
      e.stopPropagation();
      try {
        const response = await fetch(`/api/v1/apps/${selectedApp?.id}/tasks/${task.id}`, {
          method: 'PATCH',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            title: task.title,
            description: task.description,
            status: "in-progress"
          }),
        });
        
        if (response.ok) {
          // Refresh tasks to reflect new status
          fetchTasks(selectedApp?.id || 0);
        }
      } catch (error) {
        console.error('Error submitting task:', error);
      }
    };

    return (
      <div 
        onClick={() => {
          if ((column === 'inProgress' || column === 'completed') && !isEditing) {
            setSelectedTask(task)
            fetchLogs(task.id || 0);
          }}}
        className={`bg-white rounded-xl p-4 shadow-sm border border-slate-200 hover:shadow-md transition-all duration-200 ${
          column === 'inProgress' && !isEditing ? 'cursor-pointer hover:border-indigo-300' : ''
        }`}
      >
        <div className="flex items-start justify-between mb-2">
          {isEditing ? (
            <input
              type="text"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
              className="flex-1 font-semibold text-slate-800 border border-slate-300 rounded px-2 py-1 text-sm focus:outline-none focus:border-indigo-500"
              onClick={(e) => e.stopPropagation()}
            />
          ) : (
            <h4 className="font-semibold text-slate-800 flex-1">{task.title}</h4>
          )}
          
          {column === 'authoring' && (
            <div className="flex items-center gap-1 ml-2">
              {isEditing ? (
                <>
                  <button
                    onClick={handleSave}
                    className="p-1 text-green-600 hover:bg-green-50 rounded transition-colors"
                  >
                    <Check size={14} />
                  </button>
                  <button
                    onClick={handleCancel}
                    className="p-1 text-red-600 hover:bg-red-50 rounded transition-colors"
                  >
                    <X size={14} />
                  </button>
                </>
              ) : (
                <button
                  onClick={handleEdit}
                  className="p-1 text-slate-400 hover:text-slate-600 hover:bg-slate-50 rounded transition-colors"
                >
                  <Edit2 size={14} />
                </button>
              )}
            </div>
          )}
        </div>

        {isEditing ? (
          <textarea
            value={editDescription}
            onChange={(e) => setEditDescription(e.target.value)}
            className="w-full text-slate-600 text-sm border border-slate-300 rounded px-2 py-1 h-16 resize-none focus:outline-none focus:border-indigo-500"
            onClick={(e) => e.stopPropagation()}
          />
        ) : (
          <p className="text-slate-600 text-sm mb-3 line-clamp-2">{task.description}</p>
        )}
        
        {task.progress && (
          <div className="mb-3">
            <div className="flex justify-between text-xs text-slate-600 mb-1">
              <span>Progress</span>
              <span>{task.progress}%</span>
            </div>
            <div className="w-full bg-slate-200 rounded-full h-2">
              <div 
                className="bg-gradient-to-r from-indigo-500 to-purple-500 h-2 rounded-full transition-all duration-300"
                style={{ width: `${task.progress}%` }}
              />
            </div>
          </div>
        )}

        <div className="flex items-center justify-between">
          <div className="text-xs text-slate-500">
            Created: {task.created}
            {task.completedDate && (
              <span className="block">Completed: {task.completedDate}</span>
            )}
          </div>

          {column === 'authoring' && !isEditing && (
            <button
              onClick={handleSubmit}
              className="bg-gradient-to-r from-indigo-500 to-purple-500 text-white px-3 py-1 rounded-lg hover:from-indigo-600 hover:to-purple-600 transition-all duration-200 flex items-center gap-1 text-xs"
            >
              <Send size={12} />
              Submit
            </button>
          )}
        </div>
      </div>
    );
  };


  const TaskLogModal = ({ task, onClose }: { task: Task; onClose: () => void }) => (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl p-6 w-full max-w-2xl max-h-[80vh] overflow-hidden">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-xl font-semibold text-slate-800">{task.title} - Execution Log</h3>
          <button 
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 transition-colors"
          >
            ✕
          </button>
        </div>
        
        <div className="space-y-4 overflow-y-auto max-h-96">
          {logs?.messages.map((log, index) => (
            <div key={index} className="flex gap-4 p-4 bg-slate-50 rounded-xl">
              <div className="flex-shrink-0">
                {log.title === 'update' && <CheckCircle className="text-green-500" size={20} />}
                {log.title === 'tool' && <Pickaxe className="text-slate-400" size={20} />}
                {/* {log.status === 'pending' && <Clock className="text-slate-400" size={20} />} */}
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <span className="font-medium text-slate-800">{log.title == 'update' ? 'Update' : 'Tool Execution'}</span>
                  <span className="text-xs text-slate-500">{log.time}</span>
                </div>
                <p className="text-sm text-slate-600">{log.text}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );

  const AppScreen = () => (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      <div className="container mx-auto px-6 py-8">
        <div className="flex items-center gap-4 mb-8">
          <button 
            onClick={() => setCurrentView('home')}
            className="p-2 hover:bg-white/80 rounded-lg transition-colors"
          >
            <ArrowLeft className="text-slate-600" size={24} />
          </button>
          <div className="flex-1">
            <h1 className="text-3xl font-bold text-slate-800 mb-1">{selectedApp?.name}</h1>
            <p className="text-slate-600">{selectedApp?.description}</p>
          </div>
          <div className="flex gap-3">
            <button className="bg-gradient-to-r from-emerald-500 to-teal-500 text-white px-4 py-2 rounded-lg hover:from-emerald-600 hover:to-teal-600 transition-all duration-200 flex items-center gap-2" onClick={() => window.open(`/apps/${selectedApp?.id}`, '_blank')}>
              <Eye size={16} />
              Launch
            </button>
            <button className="bg-gradient-to-r from-slate-600 to-slate-700 text-white px-4 py-2 rounded-lg hover:from-slate-700 hover:to-slate-800 transition-all duration-200 flex items-center gap-2" onClick={() => downloadCode(selectedApp?.id)}>
              <Download size={16} />
              Download Code
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Authoring Column */}
          <div className="bg-white/70 backdrop-blur-sm rounded-xl p-4 border border-slate-200">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-3 h-3 bg-yellow-400 rounded-full"></div>
              <h2 className="text-lg font-semibold text-slate-800">Authoring</h2>
              <span className="bg-yellow-100 text-yellow-700 text-xs px-2 py-1 rounded-full">
                {tasks.authoring.length}
              </span>
            </div>
            <div className="space-y-3 mb-4">
              {tasks.authoring.map((task) => (
                <TaskCard key={task.id} task={task} column="authoring" />
              ))}
            </div>
            <button 
              onClick={() => {
                const newTask: Task = {
                  id: null,
                  title: 'New Task',
                  description: 'Describe what you want to build...',
                  created: new Date().toISOString().split('T')[0],
                  status: 'authoring',
                  color: 'from-yellow-400 to-yellow-500',
                  progress: null,
                  completedDate: ''
                };
                setTasks(prev => ({
                  ...prev,
                  authoring: [...prev.authoring, newTask]
                }));
              }}
              className="w-full p-3 border-2 border-dashed border-slate-300 rounded-xl text-slate-600 hover:border-slate-400 hover:text-slate-700 transition-all duration-200 flex items-center justify-center gap-2"
            >
              <Plus size={20} />
              Add Task
            </button>
          </div>

          {/* In Progress Column */}
          <div className="bg-white/70 backdrop-blur-sm rounded-xl p-4 border border-slate-200">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-3 h-3 bg-indigo-400 rounded-full"></div>
              <h2 className="text-lg font-semibold text-slate-800">In Progress</h2>
              <span className="bg-indigo-100 text-indigo-700 text-xs px-2 py-1 rounded-full">
                {tasks.inProgress.length}
              </span>
            </div>
            <div className="space-y-3">
              {tasks.inProgress.map((task) => (
                <TaskCard key={task.id} task={task} column="inProgress" />
              ))}
            </div>
          </div>

          {/* Completed Column */}
          <div className="bg-white/70 backdrop-blur-sm rounded-xl p-4 border border-slate-200">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-3 h-3 bg-green-400 rounded-full"></div>
              <h2 className="text-lg font-semibold text-slate-800">Completed</h2>
              <span className="bg-green-100 text-green-700 text-xs px-2 py-1 rounded-full">
                {tasks.completed.length}
              </span>
            </div>
            <div className="space-y-3">
              {tasks.completed.map((task) => (
                <TaskCard key={task.id} task={task} column="completed" />
              ))}
            </div>
          </div>
        </div>
      </div>

      {selectedTask && (
        <TaskLogModal 
          task={selectedTask} 
          onClose={() => setSelectedTask(null)} 
        />
      )}
    </div>
  );

  return (
    <div className="font-sans">
      {currentView === 'home' ? <AppsHomeScreen /> : <AppScreen />}
    </div>
  );
};

export default UmamiApp;