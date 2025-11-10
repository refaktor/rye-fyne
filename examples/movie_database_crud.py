#!/usr/bin/env python3

import tkinter as tk
from tkinter import ttk
import sqlite3

class Movies:
    def __init__(self, db):
        self.db = db
        self.users = []
        self.index = 0
    
    def current(self):
        if self.users and 0 <= self.index < len(self.users):
            return self.users[self.index]
        return None
    
    def next(self):
        self.index = min(self.index + 1, len(self.users) - 1)
    
    def prev(self):
        self.index = max(0, self.index - 1)
    
    def index_up(self, i):
        self.index = min(i, len(self.users) - 1) if self.users else 0
    
    def new_index(self):
        self.index = len(self.users)
    
    def is_new(self):
        return self.index >= len(self.users)
    
    def insert(self, name, score):
        cursor = self.db.cursor()
        cursor.execute("INSERT INTO movies (name, score) VALUES (?, ?)", (name, score))
        self.db.commit()
    
    def update(self, id, name, score):
        cursor = self.db.cursor()
        cursor.execute("UPDATE movies SET name = ?, score = ? WHERE id = ?", (name, score, id))
        self.db.commit()
    
    def delete(self, id):
        cursor = self.db.cursor()
        cursor.execute("DELETE FROM movies WHERE id = ?", (id,))
        self.db.commit()
    
    def refresh(self):
        cursor = self.db.cursor()
        cursor.execute("SELECT id, name, score FROM movies ORDER BY id")
        self.users = cursor.fetchall()
        self.index_up(self.index)


def main():
    # Database connection
    db = sqlite3.connect("movies.db")
    
    # Movies context
    movies = Movies(db)
    
    # GUI setup
    root = tk.Tk()
    root.title("My Movie Database [MYMDb]")
    root.geometry("300x250")
    root.resizable(False, False)
    
    # Entry widgets
    name_entry = ttk.Entry(root)
    score_entry = ttk.Entry(root)
    
    def update_entries():
        current = movies.current()
        if current:
            name_entry.delete(0, tk.END)
            name_entry.insert(0, current[1])  # name
            score_entry.delete(0, tk.END) 
            score_entry.insert(0, str(current[2]))  # score
        else:
            name_entry.delete(0, tk.END)
            score_entry.delete(0, tk.END)
    
    # Button functions
    def prev_clicked():
        movies.prev()
        update_entries()
    
    def next_clicked():
        movies.next()
        update_entries()
    
    def new_clicked():
        name_entry.delete(0, tk.END)
        score_entry.delete(0, tk.END)
        movies.new_index()
    
    def save_clicked():
        name = name_entry.get()
        score = int(score_entry.get()) if score_entry.get() else 0
        
        if movies.is_new():
            movies.insert(name, score)
        else:
            current = movies.current()
            if current:
                movies.update(current[0], name, score)  # current[0] is id
        
        movies.refresh()
        update_entries()
    
    def delete_clicked():
        current = movies.current()
        if current:
            movies.delete(current[0])  # current[0] is id
            movies.refresh()
            update_entries()
    
    # Buttons
    nav_frame = ttk.Frame(root)
    nav_frame.pack(pady=5)
    
    prev_btn = ttk.Button(nav_frame, text="Previous", command=prev_clicked)
    prev_btn.pack(side=tk.LEFT, padx=5)
    
    next_btn = ttk.Button(nav_frame, text="Next", command=next_clicked)
    next_btn.pack(side=tk.LEFT, padx=5)
    
    # Labels and entries
    ttk.Label(root, text="Name:").pack(pady=(10, 2))
    name_entry.pack(pady=2)
    
    ttk.Label(root, text="Score:").pack(pady=(10, 2))
    score_entry.pack(pady=2)
    
    # Action buttons
    action_frame = ttk.Frame(root)
    action_frame.pack(pady=15)
    
    new_btn = ttk.Button(action_frame, text="New", command=new_clicked)
    new_btn.pack(side=tk.LEFT, padx=5)
    
    save_btn = ttk.Button(action_frame, text="Save", command=save_clicked)
    save_btn.pack(side=tk.LEFT, padx=5)
    
    delete_btn = ttk.Button(action_frame, text="Delete", command=delete_clicked)
    delete_btn.pack(side=tk.LEFT, padx=5)
    
    # Initialize
    movies.refresh()
    update_entries()
    
    # Handle window close
    def on_closing():
        db.close()
        root.destroy()
    
    root.protocol("WM_DELETE_WINDOW", on_closing)
    
    # Show window
    root.mainloop()


if __name__ == "__main__":
    main()


# Prerequisites movies.db
#
#    CREATE TABLE IF NOT EXISTS movies (
#        id INTEGER PRIMARY KEY AUTOINCREMENT,
#        name VARCHAR(100) NOT NULL,
#        score INTEGER NOT NULL
#    );
#
#
#    INSERT INTO movies (name, score) VALUES 
#        ('Stalker', 9), ('Gattaca', 8), 
#        ('Starship Troopers', 7);
