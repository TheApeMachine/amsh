# filename: plot_stocks.py
import os

# Ensure yfinance and matplotlib are installed
os.system('pip install yfinance matplotlib')

import yfinance as yf
import matplotlib.pyplot as plt
from datetime import datetime

# Get current date
today = datetime.now()
start_of_year = datetime(today.year, 1, 1)

# Fetch stock data
nvda = yf.download('NVDA', start=start_of_year)
tsla = yf.download('TSLA', start=start_of_year)

# Calculate YTD percentage change using iloc
nvda_ytd_change = (nvda['Close'].iloc[-1] - nvda['Close'].iloc[0]) / nvda['Close'].iloc[0] * 100
tsla_ytd_change = (tsla['Close'].iloc[-1] - tsla['Close'].iloc[0]) / tsla['Close'].iloc[0] * 100

# Prepare the data for plotting
stocks = ['NVIDIA', 'Tesla']
changes = [nvda_ytd_change, tsla_ytd_change]

# Plotting
plt.bar(stocks, changes, color=['blue', 'orange'])
plt.title('YTD Stock Price Change for NVDA and TSLA')
plt.ylabel('Percentage Change (%)')
plt.ylim(min(changes) - 10, max(changes) + 10)  # Adjust y-axis for better view
plt.grid(axis='y')

# Save the plot to a file
plt.savefig('plot.png')
plt.close()