import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';
import { Centrifuge } from 'centrifuge';
import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';

function App() {
  const [state, setState] = useState({
    leaders: [],
    prevOrder: {},
    // lastSeq: 0,
    // epoch: '',
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
        // // Check for epoch changes
        // if (prevState.epoch && data.epoch !== prevState.epoch) {
        //   // The epoch changed, so reset the state or handle accordingly.
        //   return {
        //     leaders: data.leaders,
        //     prevOrder: {},
        //     lastSeq: data.seq,
        //     epoch: data.epoch,
        //     highlights: {},
        //   };
        // }
        
        // // Ignore messages with an older or equal sequence
        // if (data.seq <= prevState.lastSeq) return prevState;
    
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
          // lastSeq: data.seq,
          highlights: { ...prevState.highlights, ...newHighlights },
          // epoch: data.epoch, // update epoch if not already set
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
      <h1 className="mb-4">Real-time Leaderboard with Centrifugo</h1>
      <div className="card">
        <div className="card-body">
          <table className="table table-striped">
            <thead>
              <tr>
                <th scope="col">Rank</th>
                <th scope="col">Name</th>
                <th scope="col">Score</th>
              </tr>
            </thead>
            <tbody>
              {state.leaders.map((leader, index) => (
                <motion.tr key={leader.name} layout>
                  <td className={state.highlights[leader.name] || ''}>{index + 1}</td>
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
