import { useEffect, useState } from "react"
import { getTimeRemaining } from './utils.js';

const Cooldown = ({ endDate }) => {
    const [time, setTime] = useState<any>({ hours: 0, minutes: 0, seconds: 0});
    const { hours, minutes, seconds } = time; 
    useEffect(() => {
        const interval = setInterval(() => setTime(getTimeRemaining(new Date(endDate * 1000))), 1000);
        return () => {
            clearInterval(interval)
        }
    }, [endDate])
    
    return (
        <span style={{ fontSize: "12px"}}>{String(hours).padStart(2, "0")}:{String(minutes).padStart(2, "0")}:{String(seconds).padStart(2, "0")}</span>
    );
}

export { Cooldown }