"use client";

import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Zap, CheckCircle, Loader2, Clock, Plus, Server, Activity, Cpu, Box, LayoutDashboard } from 'lucide-react';

interface TaskUpdate {
  id: string;
  status: string;
  time: string;
  worker?: string;
  message?: string;
}

export default function TaskDashboard() {
  const [tasks, setTasks] = useState<TaskUpdate[]>([]);
  const [connected, setConnected] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // WebSocket Connection
  useEffect(() => {
    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8081/ws';
    const socket = new WebSocket(wsUrl);
    socket.onopen = () => setConnected(true);
    socket.onclose = () => setConnected(false);
    socket.onmessage = (event) => {
      const update: TaskUpdate = JSON.parse(event.data);
      setTasks((prev) => {
        const exists = prev.find(t => t.id === update.id);
        if (exists) return prev.map(t => t.id === update.id ? update : t);
        return [update, ...prev].slice(0, 12); // Increased buffer for grid
      });
    };
    return () => socket.close();
  }, []);

  const handleSubmitTask = async () => {
    setIsSubmitting(true);
    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      console.log('Submitting task to:', `${apiUrl}/submit`);
      
      const response = await fetch(`${apiUrl}/submit`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        // Prevent browser from following redirects or caching
        cache: 'no-cache',
        redirect: 'follow',
      });
      
      if (!response.ok) {
        console.error('Failed to submit task:', response.status, response.statusText);
        alert(`Failed to submit task: ${response.status} ${response.statusText}`);
      } else {
        const data = await response.json();
        console.log('Task submitted successfully:', data);
      }
    } catch (error) {
      console.error('Error submitting task:', error);
      alert(`Error: ${error instanceof Error ? error.message : 'Failed to submit task'}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Define our cluster nodes
  const nodes = ['manager', 'worker-1', 'worker-2', 'worker-3'];

  return (
    <div className="min-h-screen bg-[#09090b] text-slate-300 font-sans p-6 lg:p-12 selection:bg-blue-500/30">
      {/* Background Ambient Glows */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <div className="absolute top-1/4 left-1/4 w-[600px] h-[600px] bg-blue-600/[0.02] blur-[120px] rounded-full" />
        <div className="absolute bottom-1/4 right-1/4 w-[600px] h-[600px] bg-purple-600/[0.02] blur-[120px] rounded-full" />
      </div>

      <div className="relative z-10 max-w-7xl mx-auto">
        {/* System Header */}
        <header className="flex flex-col sm:flex-row items-center justify-between mb-12 gap-8 bg-[#111114] p-8 rounded-[40px] border border-white/5 shadow-2xl">
          <div className="flex items-center gap-6">
            <div className="w-16 h-16 bg-white text-black rounded-[24px] flex items-center justify-center shadow-2xl ring-8 ring-white/5">
              <LayoutDashboard size={32} />
            </div>
            <div>
              <h1 className="text-4xl font-black text-white tracking-tighter">Cluster Control</h1>
              <div className="flex items-center gap-2 mt-1">
                <div className={`w-2 h-2 rounded-full ${connected ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`} />
                <span className="text-[11px] font-black uppercase tracking-[0.2em] text-slate-500">
                  {connected ? 'Network Active' : 'System Offline'}
                </span>
              </div>
            </div>
          </div>

          <button 
            type="button"
            onClick={handleSubmitTask}
            disabled={isSubmitting || !connected}
            className="group px-10 py-5 bg-white text-black rounded-[24px] font-black text-lg hover:bg-slate-200 active:scale-[0.98] transition-all disabled:opacity-20 flex items-center gap-4 shadow-xl shadow-white/5"
          >
            {isSubmitting ? <Loader2 className="animate-spin" size={24} /> : <><Plus strokeWidth={3} size={24} className="group-hover:rotate-90 transition-transform" /> <span>Deploy Node</span></>}
          </button>
        </header>

        {/* Workers Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
          {nodes.map((nodeId) => {
            const nodeTasks = tasks.filter(t => (t.worker || 'manager').toLowerCase() === nodeId.toLowerCase());
            
            return (
              <div key={nodeId} className="flex flex-col gap-5">
                {/* Node Label */}
                <div className="flex items-center justify-between px-6 py-2 text-[11px] font-black uppercase tracking-[0.3em] text-slate-600 bg-white/[0.02] rounded-full border border-white/5">
                  <div className="flex items-center gap-2">
                    <Cpu size={14} className={nodeTasks.length > 0 ? "text-blue-500" : ""} />
                    <span>{nodeId}</span>
                  </div>
                  <span className="text-slate-700">{nodeTasks.length}</span>
                </div>

                {/* Task Stack */}
                <div className="flex-1 bg-black/20 border border-white/5 rounded-[40px] p-4 min-h-[500px] backdrop-blur-sm">
                  <AnimatePresence mode="popLayout">
                    {nodeTasks.length === 0 ? (
                      <div className="h-full flex flex-col items-center justify-center opacity-10 py-20" key="empty">
                        <Box size={40} strokeWidth={1.5} />
                        <span className="text-[10px] mt-4 font-black uppercase tracking-[0.4em]">Standby</span>
                      </div>
                    ) : (
                      <div className="space-y-4">
                        {nodeTasks.map((task) => (
                          <motion.div
                            layout
                            initial={{ opacity: 0, y: 20, scale: 0.9 }}
                            animate={{ opacity: 1, y: 0, scale: 1 }}
                            exit={{ opacity: 0, scale: 0.9, transition: { duration: 0.2 } }}
                            key={task.id}
                            className="bg-[#1a1a1e] border border-white/[0.08] p-6 rounded-[32px] hover:border-white/20 transition-all shadow-xl"
                          >
                            <div className="flex items-center justify-between mb-4">
                              <div className={`p-2.5 rounded-xl bg-white/5 ${getStatusColor(task.status)}`}>
                                {getStatusIcon(task.status)}
                              </div>
                              <span className="text-[10px] font-mono text-slate-600 font-bold tracking-tighter">
                                #{task.id.slice(0, 6)}
                              </span>
                            </div>

                            <div className="space-y-2 mb-6">
                              <h4 className="text-white font-bold text-sm tracking-tight">
                                {task.status.toUpperCase() === 'COMPLETED' ? 'Task Finalized' : 'In Progress'}
                              </h4>
                              <p className="text-[12px] text-slate-400 font-semibold leading-relaxed">
                                {task.message || 'Processing instructions...'}
                              </p>
                            </div>

                            <div className="flex items-center justify-between gap-3 text-[10px] font-black text-slate-600 uppercase tracking-widest mb-3">
                              <span>{task.time}</span>
                              <span className={task.status.toUpperCase() === 'COMPLETED' ? "text-emerald-500" : "text-blue-500"}>
                                {getProgress(task.status)}%
                              </span>
                            </div>

                            <div className="h-1.5 w-full bg-black rounded-full overflow-hidden shadow-inner">
                              <motion.div 
                                initial={{ width: 0 }}
                                animate={{ width: `${getProgress(task.status)}%` }}
                                transition={{ duration: 1, ease: "circOut" }}
                                className={`h-full ${getBarColor(task.status)}`}
                              />
                            </div>
                          </motion.div>
                        ))}
                      </div>
                    )}
                  </AnimatePresence>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

// UI Mapping Helpers
const getStatusIcon = (s: string) => {
  const status = s.toUpperCase();
  if (status === 'COMPLETED') return <CheckCircle size={18} />;
  if (status === 'PROCESSING') return <Loader2 size={18} className="animate-spin" />;
  return <Clock size={18} />;
};

const getStatusColor = (s: string) => {
  const status = s.toUpperCase();
  if (status === 'COMPLETED') return 'text-emerald-400';
  if (status === 'PROCESSING') return 'text-blue-400';
  return 'text-amber-400';
};

const getBarColor = (s: string) => {
  const status = s.toUpperCase();
  if (status === 'COMPLETED') return 'bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.3)]';
  if (status === 'PROCESSING') return 'bg-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.3)]';
  return 'bg-amber-500 shadow-[0_0_10px_rgba(245,158,11,0.3)]';
};

const getProgress = (s: string) => {
  const status = s.toUpperCase();
  if (status === 'COMPLETED') return 100;
  if (status === 'PROCESSING') return 65;
  return 15;
};
