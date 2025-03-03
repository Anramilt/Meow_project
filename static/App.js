import React, {useState} from "react";
import ReactDOM from "react-dom";
import "./static/styles.css";
function App() {
    const [squares, setSquares] = useState(Array(9).fill(null));
    const [xIsNext, setIsNext] = useState(true);
    
    //Функция проверяет на каждом ходе, есть ли победитель
    const calculateWinner = (squares) => {
        //внутренние выигрышные комбинации
        const lines = [
          [0, 1, 2],
          [3, 4, 5],
          [6, 7, 8],
          [0, 3, 6],
          [1, 4, 7],
          [2, 5, 8],
          [0, 4, 8],
          [2, 4, 6],
        ];
        for (let i=0; i< lines.length; i++){
            const [a, b, c] =lines[i];
            if (squares[a] && squares[a] === squares[b] && squares[a] === squares[c]){
                return squares[a];
            }
        }
        return null;
    };

    const winner = calculateWinner(squares);

    //вызывается при клике на игровое поле
    const handleClick = (i) => {
        const squaresCopy = [...squares];
        if (winner || squaresCopy[i]) return;
        squaresCopy[i] = xIsNext ? "X" : "O";
        setSquares(squaresCopy);
        setIsNext(!xIsNext);
    }
    
    const renderSquare = (i) => {
        return (
            <button className="square" onClick={() => handleClick(i)}>
                {squares[i]}
            </button>
        );
    };

    const status = winner
    ? `Winner: ${winner}` : `Next player: ${xIsNext ? "X" : "O"}`;

    return(
        <div className="game-field">
      <div className="status">{status}</div>
      <div className="board-row">
        {renderSquare(0)}
        {renderSquare(1)}
        {renderSquare(2)}
      </div>
      <div className="board-row">
        {renderSquare(3)}
        {renderSquare(4)}
        {renderSquare(5)}
      </div>
      <div className="board-row">
        {renderSquare(6)}
        {renderSquare(7)}
        {renderSquare(8)}
      </div>
      <div className="button-container">
                <button className="button" onClick={() => setSquares(Array(9).fill(null))}>
                    Начать заново
                </button>
            </div>
    </div>
    )
    
}
export default App;