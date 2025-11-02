import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Plus, Trash2, FileText, ChevronDown, ChevronUp } from 'lucide-react';
import api from '../lib/api';
import toast from 'react-hot-toast';
import { Note } from '../types';
import NewNoteModal from '../components/NewNoteModal';

export default function ProjectDetail() {
  const { projectId } = useParams();
  const navigate = useNavigate();
  const [notes, setNotes] = useState<Note[]>([]);
  const [projectName, setProjectName] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [expandedNoteId, setExpandedNoteId] = useState<string | null>(null);

  useEffect(() => {
    fetchNotes();
  }, [projectId]);

  const fetchNotes = async () => {
    try {
      const { data } = await api.get(`/projects/${projectId}`);
      setNotes(data);
    } catch (error) {
      toast.error('Failed to load notes');
      console.error('Fetch notes error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateNote = async (title: string, content: string) => {
    try {
      const { data } = await api.post('/notes', {
        project_id: projectId,
        title,
        content
      });
      setNotes([...notes, data]);
      setIsModalOpen(false);
      toast.success('Note created successfully!');
    } catch (error) {
      toast.error('Failed to create note');
      console.error('Create note error:', error);
    }
  };

  const handleDeleteNote = async (noteId: string) => {
    if (!confirm('Are you sure you want to delete this note?')) return;

    try {
      await api.delete(`/notes/${noteId}`);
      setNotes(notes.filter(note => note.id !== noteId));
      toast.success('Note deleted successfully!');
    } catch (error) {
      toast.error('Failed to delete note');
      console.error('Delete note error:', error);
    }
  };

  const toggleNoteExpansion = (noteId: string) => {
    setExpandedNoteId(expandedNoteId === noteId ? null : noteId);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-gray-600">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <button
                onClick={() => navigate('/projects')}
                className="text-gray-600 hover:text-gray-900 transition-colors"
              >
                <ArrowLeft size={24} />
              </button>
              <h1 className="text-2xl font-bold text-gray-900">
                {projectName || 'Project Notes'}
              </h1>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-6">
          <button
            onClick={() => setIsModalOpen(true)}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            <Plus size={20} />
            Add Note
          </button>
        </div>

        {notes.length === 0 ? (
          <div className="text-center py-12">
            <FileText size={48} className="mx-auto text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No notes yet</h3>
            <p className="text-gray-600">Add your first note to this project</p>
          </div>
        ) : (
          <div className="space-y-4">
            {notes.map((note) => {
              const isExpanded = expandedNoteId === note.id;
              return (
                <div
                  key={note.id}
                  className="bg-white rounded-lg shadow-sm border border-gray-200 p-6"
                >
                  <div className="flex items-start justify-between mb-3">
                    <h3 className="text-lg font-semibold text-gray-900 flex-1">
                      {note.title}
                    </h3>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => toggleNoteExpansion(note.id)}
                        className="text-gray-500 hover:text-gray-700 transition-colors"
                      >
                        {isExpanded ? <ChevronUp size={20} /> : <ChevronDown size={20} />}
                      </button>
                      <button
                        onClick={() => handleDeleteNote(note.id)}
                        className="text-red-500 hover:text-red-700 transition-colors"
                      >
                        <Trash2 size={18} />
                      </button>
                    </div>
                  </div>

                  <div className={`text-gray-600 ${isExpanded ? '' : 'line-clamp-3'}`}>
                    {note.content || 'No content'}
                  </div>

                  {!isExpanded && note.content && note.content.length > 150 && (
                    <button
                      onClick={() => toggleNoteExpansion(note.id)}
                      className="text-blue-600 hover:text-blue-700 text-sm mt-2"
                    >
                      Read more
                    </button>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </main>

      <NewNoteModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateNote}
      />
    </div>
  );
}
