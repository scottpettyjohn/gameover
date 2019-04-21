import {Component, OnDestroy, OnInit} from '@angular/core';
import {SocketService} from "../socket.service";

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']
})
export class HomeComponent implements OnInit, OnDestroy{
  public hour: string;
  public min: string;
  public sec: string;

  public constructor(private socket: SocketService) {
    this.hour = '00';
    this.min = '00';
    this.sec = '00';
  }

  ngOnInit(): void {
    this.socket.getEventListener().subscribe( event => {
      if(event.type == 'message') {
        let data = event.data;
        if(data.Type === 3) {
          this.parseTime(data.Data)
        }
      }
      if(event.type == 'close') {
        // TODO
      }
      if(event.type == 'open') {
        // TODO
      }
    })
  }

  parseTime(totalseconds: number): void {
    const hourNum = Math.floor(totalseconds / 3600);
    const remainderMins = totalseconds % 3600;
    const minNum  = Math.floor(remainderMins / 60);
    const secNum = Math.floor(remainderMins % 60);
    this.hour = this.numToString(hourNum);
    this.min = this.numToString(minNum);
    this.sec = this.numToString(secNum);
  }

  numToString(num: number): string {
    if(num >= 10) {
      return ''+num;
    } else {
      return '0'+num;
    }
  }

  ngOnDestroy(): void {
    this.socket.close();
  }
}
