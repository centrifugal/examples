import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';
import { Centrifuge } from 'centrifuge';
import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';

function App() {
  const [state, setState] = useState({
    leaders: [],
    prevOrder: {},
    highlights: {},
  });

  useEffect(() => {
    const centrifuge = new Centrifuge("ws://localhost:8000/connection/websocket");
    const sub = centrifuge.newSubscription("leaderboard", {
      delta: 'fossil',
      since: {}
    });

    sub.on('publication', (message) => {
      const data = message.data;
 
      setState(prevState => {    
        const newHighlights = {};
        const newLeaders = data.leaders.map((leader, index) => {
          let highlightClass = "";
          const prevRank = prevState.prevOrder[leader.name];
          if (prevRank !== undefined) {
            if (prevRank > index) {
              highlightClass = "highlight-up";
            } else if (prevRank < index) {
              highlightClass = "highlight-down";
            }
          }
          if (highlightClass) {
            newHighlights[leader.name] = highlightClass;
            setTimeout(() => {
              setState(current => ({
                ...current,
                highlights: { ...current.highlights, [leader.name]: "" }
              }));
            }, 1000);
          }
          return leader;
        });
    
        const newOrder = {};
        newLeaders.forEach((leader, index) => {
          newOrder[leader.name] = index;
        });
    
        return {
          ...prevState,
          leaders: newLeaders,
          prevOrder: newOrder,
          highlights: { ...prevState.highlights, ...newHighlights },
        };
      });
    });

    centrifuge.connect();
    sub.subscribe();

    return () => {
      sub.unsubscribe();
      centrifuge.disconnect();
    };
  }, []);

  return (
    <div className="container mt-5">
      <div className="card">
        <div className="card-header">Real-time Leaderboard with Centrifugo</div>
        <div className="card-body">
          <table className="table table-striped">
            <thead>
              <tr>
                <th scope="col" className="rank-col">Rank</th>
                <th scope="col">Name</th>
                <th scope="col">Score</th>
              </tr>
            </thead>
            <tbody>
              {state.leaders.map((leader, index) => (
                <motion.tr key={leader.name} layout>
                  <td className={`rank-col ${state.highlights[leader.name] || ''}`}>{index + 1}</td>
                  <td className={state.highlights[leader.name] || ''}>{leader.name}</td>
                  <td className={state.highlights[leader.name] || ''}>{leader.score}</td>
                </motion.tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export default App;