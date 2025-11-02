export interface User {
  id: string;
  email: string;
  name: string;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  note_count?: number;
}

export interface Note {
  id: string;
  project_id: string;
  title: string;
  content: string;
  created_at: string;
}
